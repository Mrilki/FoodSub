package com.example.auth.service;

import com.example.auth.dto.AuthRequest;
import com.example.auth.dto.AuthResponse;
import com.example.auth.entity.UserEntity;
import com.example.auth.entity.UserProfileEntity;
import com.example.auth.entity.UserSessionEntity;
import com.example.auth.enums.UserRole;
import com.example.auth.repository.UserProfileRepository;
import com.example.auth.repository.UserRepository;
import com.example.auth.repository.UserSessionRepository;
import com.example.auth.security.JwtTokenProvider;
import com.example.auth.security.TokenHashUtil;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.security.authentication.AuthenticationManager;
import org.springframework.security.authentication.UsernamePasswordAuthenticationToken;
import org.springframework.security.core.Authentication;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.slf4j.MDC;

import java.time.LocalDateTime;
import java.util.HashMap;
import java.util.Map;
import java.util.UUID;

@Slf4j
@Service
@RequiredArgsConstructor
public class AuthService {

    private final UserRepository userRepository;
    private final UserProfileRepository profileRepository;
    private final UserSessionRepository sessionRepository;
    private final PasswordEncoder passwordEncoder;
    private final JwtTokenProvider tokenProvider;
    private final AuthenticationManager authenticationManager;
    private final KafkaTemplate<String, Object> kafkaTemplate;

    @Transactional
    public AuthResponse register(AuthRequest request) {
        if (userRepository.existsByEmail(request.getEmail())) {
            throw new RuntimeException("Email already exists");
        }
        if (request.getPassword().length() > 72) {
            throw new IllegalArgumentException("Password is too long (max 72 characters)");
        }
        String hashedPassword = passwordEncoder.encode(request.getPassword());

        UserEntity user = UserEntity.builder()
                .email(request.getEmail())
                .passwordHash(hashedPassword)
                .role(UserRole.USER)
                .active(true)
                .build();
        userRepository.save(user);

        UserProfileEntity profile = UserProfileEntity.builder()
                .user(user)
                .build();
        profileRepository.save(profile);

        Authentication authentication = authenticationManager.authenticate(
                new UsernamePasswordAuthenticationToken(request.getEmail(), request.getPassword())
        );

        String accessToken = tokenProvider.generateAccessToken(authentication);
        String refreshToken = tokenProvider.generateRefreshToken(authentication);

        saveSession(user, refreshToken);

        sendKafkaEvent("user.created", user);

        log.info("Register attempt: email={}, passwordLength={}",
                request.getEmail(),
                request.getPassword() != null ? request.getPassword().length() : "null");
        return AuthResponse.of(accessToken, refreshToken);
    }

    @Transactional
    public AuthResponse login(AuthRequest request) {
        Authentication authentication = authenticationManager.authenticate(
                new UsernamePasswordAuthenticationToken(request.getEmail(), request.getPassword())
        );

        UserEntity user = userRepository.findByEmail(request.getEmail())
                .orElseThrow(() -> new RuntimeException("User not found"));

        if (!user.isActive()) {
            throw new RuntimeException("User account is blocked");
        }

        String accessToken = tokenProvider.generateAccessToken(authentication);
        String refreshToken = tokenProvider.generateRefreshToken(authentication);

        saveSession(user, refreshToken);

        return AuthResponse.of(accessToken, refreshToken);
    }

    @Transactional
    public AuthResponse refreshTokens(String refreshToken) {
        if (!tokenProvider.validateToken(refreshToken)) {
            throw new RuntimeException("Invalid refresh token");
        }

        String tokenHash = hashToken(refreshToken);
        UserSessionEntity session = sessionRepository.findByRefreshTokenHash(tokenHash)
                .orElseThrow(() -> new RuntimeException("Session not found"));

        if (session.getIsRevoked()) {
            throw new RuntimeException("Token has been revoked");
        }

        if (session.getExpiresAt().isBefore(LocalDateTime.now())) {
            throw new RuntimeException("Token expired");
        }

        UserEntity user = session.getUser();
        Authentication authentication = authenticationManager.authenticate(
                new UsernamePasswordAuthenticationToken(user.getEmail(), null)
        );

        session.setIsRevoked(true);
        sessionRepository.save(session);

        String newAccessToken = tokenProvider.generateAccessToken(authentication);
        String newRefreshToken = tokenProvider.generateRefreshToken(authentication);

        saveSession(user, newRefreshToken);

        return AuthResponse.of(newAccessToken, newRefreshToken);
    }

    @Transactional
    public void logout(String refreshToken) {
        String tokenHash = hashToken(refreshToken);
        sessionRepository.findByRefreshTokenHash(tokenHash)
                .ifPresent(session -> {
                    session.setIsRevoked(true);
                    sessionRepository.save(session);
                });
    }

    private void saveSession(UserEntity user, String refreshToken) {
        UserSessionEntity session = UserSessionEntity.builder()
                .user(user)
                .refreshTokenHash(hashToken(refreshToken))
                .expiresAt(LocalDateTime.now().plusSeconds(2592000))
                .isRevoked(false)
                .build();
        sessionRepository.save(session);
    }

    private String hashToken(String token) {
        return TokenHashUtil.hash(token);
    }

    @Value("${kafka.enabled:true}")
    private boolean kafkaEnabled;

    @Value("${kafka.topic.user-events:user.events}")
    private String kafkaTopicUserEvents;

    private void sendKafkaEvent(String eventType, UserEntity user) {
        if (!kafkaEnabled) {
            log.debug("Kafka disabled, skipping event: {}", eventType);
            return;
        }

        try {
            String traceId = MDC.get("trace_id");
            if (traceId == null) {
                traceId = UUID.randomUUID().toString();
            }

            Map<String, Object> payload = new HashMap<>();
            payload.put("user_id", user.getId().toString());
            payload.put("email", user.getEmail());
            payload.put("created_at", user.getCreatedAt());

            Map<String, Object> event = new HashMap<>();
            event.put("trace_id", traceId);
            event.put("timestamp", LocalDateTime.now());
            event.put("event_type", eventType);
            event.put("payload", payload);

            kafkaTemplate.send(kafkaTopicUserEvents, event);
            log.info("Kafka event sent: {} (trace_id={})", eventType, traceId);

        } catch (Exception e) {
            log.warn("Failed to send Kafka event {}: {}", eventType, e.getMessage());
        }
    }
}