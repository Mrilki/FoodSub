package com.example.auth.service;

import com.example.auth.dto.AdminUserListResponse;
import com.example.auth.dto.AdminUserResponse;
import com.example.auth.dto.UpdateRoleRequest;
import com.example.auth.entity.UserEntity;
import com.example.auth.entity.UserProfileEntity;
import com.example.auth.repository.UserRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.PageRequest;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.UUID;

@Service
@RequiredArgsConstructor
public class AdminService {

    private final UserRepository userRepository;

    @Transactional(readOnly = true)
    public AdminUserListResponse getAllUsers(int page, int size) {
        Pageable pageable = PageRequest.of(page, size);
        Page<UserEntity> userPage = userRepository.findAll(pageable);

        var content = userPage.getContent().stream()
                .map(this::mapToAdminResponse)
                .toList();

        return new AdminUserListResponse(
                content,
                userPage.getNumber(),
                userPage.getSize(),
                userPage.getTotalElements(),
                userPage.getTotalPages()
        );
    }

    @Transactional(readOnly = true)
    public AdminUserResponse getUserById(UUID id) {
        UserEntity user = userRepository.findById(id)
                .orElseThrow(() -> new RuntimeException("User not found"));
        return mapToAdminResponse(user);
    }

    @Transactional
    public void updateUserRole(UUID id, UpdateRoleRequest request) {
        UserEntity user = userRepository.findById(id)
                .orElseThrow(() -> new RuntimeException("User not found"));

        if (user.getRole().name().equals("ADMIN") && !request.getRole().name().equals("ADMIN")) {
            long adminCount = userRepository.countByRoleAndActiveTrue(com.example.auth.enums.UserRole.ADMIN);
            if (adminCount <= 1) {
                throw new RuntimeException("Cannot remove the last admin");
            }
        }

        user.setRole(request.getRole());
        userRepository.save(user);
    }

    private AdminUserResponse mapToAdminResponse(UserEntity user) {
        UserProfileEntity profile = user.getProfile();
        AdminUserResponse.UserProfileDTO profileDTO = null;

        if (profile != null) {
            profileDTO = new AdminUserResponse.UserProfileDTO(
                    profile.getFullName(),
                    profile.getPhone(),
                    profile.getAddress(),
                    profile.getCity()
            );
        }

        return new AdminUserResponse(
                user.getId(),
                user.getEmail(),
                user.getRole(),
                user.isActive(),
                user.getCreatedAt(),
                user.getUpdatedAt(),
                profileDTO
        );
    }
}
