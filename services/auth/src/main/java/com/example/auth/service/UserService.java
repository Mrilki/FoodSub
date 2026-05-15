package com.example.auth.service;

import com.example.auth.dto.UserProfileRequest;
import com.example.auth.dto.UserResponse;
import com.example.auth.entity.UserEntity;
import com.example.auth.entity.UserProfileEntity;
import com.example.auth.repository.UserProfileRepository;
import com.example.auth.repository.UserRepository;
import com.example.auth.repository.UserSessionRepository;
import com.example.auth.security.JwtTokenProvider;
import lombok.RequiredArgsConstructor;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.UUID;

@Service
@RequiredArgsConstructor
public class UserService {

    private final UserRepository userRepository;
    private final UserProfileRepository profileRepository;
    private final JwtTokenProvider tokenProvider;
    private final PasswordEncoder passwordEncoder;
    private final UserSessionRepository sessionRepository;

    @Transactional(readOnly = true)
    public UserResponse getMe(String token) {
        String userIdStr = tokenProvider.getUserIdFromToken(token);
        UUID userId = UUID.fromString(userIdStr);

        UserEntity user = userRepository.findById(userId)
                .orElseThrow(() -> new RuntimeException("User not found"));

        return new UserResponse(
                user.getId(),
                user.getEmail(),
                user.getRole(),
                user.isActive(),
                user.getCreatedAt()
        );
    }

    @Transactional
    public void updateProfile(String token, UserProfileRequest request) {
        String userId = tokenProvider.getUserIdFromToken(token);
        UserProfileEntity profile = profileRepository.findByUserId(UUID.fromString(userId))
                .orElseThrow(() -> new RuntimeException("Profile not found"));

        if (request.getFullName() != null) profile.setFullName(request.getFullName());
        if (request.getPhone() != null) profile.setPhone(request.getPhone());
        if (request.getAddress() != null) profile.setAddress(request.getAddress());
        if (request.getCity() != null) profile.setCity(request.getCity());

        profileRepository.save(profile);
    }

    @Transactional
    public void deleteMe(String token) {
        String userId = tokenProvider.getUserIdFromToken(token);
        UserEntity user = userRepository.findById(UUID.fromString(userId))
                .orElseThrow(() -> new RuntimeException("User not found"));
        user.setActive(false);

        userRepository.save(user);
    }

    @Transactional
    public void changePassword(String token, String oldPassword, String newPassword) {
        String userId = tokenProvider.getUserIdFromToken(token);
        UserEntity user = userRepository.findById(UUID.fromString(userId))
                .orElseThrow(() -> new RuntimeException("User not found"));

        if (!passwordEncoder.matches(oldPassword, user.getPasswordHash())) {
            throw new RuntimeException("Old password is incorrect");
        }

        user.setPasswordHash(passwordEncoder.encode(newPassword));
        userRepository.save(user);

        sessionRepository.deleteByUserId(user.getId());
    }
}