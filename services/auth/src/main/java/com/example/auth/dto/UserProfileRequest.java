package com.example.auth.dto;

import lombok.Data;

@Data
public class UserProfileRequest {
    private String fullName;
    private String phone;
    private String address;
    private String city;
}
