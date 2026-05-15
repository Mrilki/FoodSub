package com.example.auth.dto;

import com.example.auth.enums.UserRole;
import lombok.AllArgsConstructor;
import lombok.Data;
import java.time.LocalDateTime;
import java.util.UUID;

@Data
@AllArgsConstructor
public class AdminUserResponse {
    private UUID id;
    private String email;
    private UserRole role;
    private Boolean isActive;
    private LocalDateTime createdAt;
    private LocalDateTime updatedAt;


    private UserProfileDTO profile;

    @Data
    @AllArgsConstructor
    public static class UserProfileDTO {
        private String fullName;
        private String phone;
        private String address;
        private String city;
    }
}