package com.example.auth.controller;

import com.example.auth.dto.PasswordChangeRequest;
import com.example.auth.dto.UserProfileRequest;
import com.example.auth.dto.UserResponse;
import com.example.auth.service.UserService;
import jakarta.servlet.http.HttpServletRequest;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/api/v1/users")
@RequiredArgsConstructor
public class UserController {

    private final UserService userService;

    @GetMapping("/me")
    public ResponseEntity<UserResponse> getMe(HttpServletRequest request) {
        String token = request.getHeader("Authorization").substring(7);
        return ResponseEntity.ok(userService.getMe(token));
    }

    @PutMapping("/me")
    public ResponseEntity<Void> updateMe(
            HttpServletRequest request,
            @RequestBody UserProfileRequest profileRequest) {
        String token = request.getHeader("Authorization").substring(7);
        userService.updateProfile(token, profileRequest);
        return ResponseEntity.ok().build();
    }

    @DeleteMapping("/me")
    public ResponseEntity<Void> deleteMe(HttpServletRequest request) {
        String token = request.getHeader("Authorization").substring(7);
        userService.deleteMe(token);
        return ResponseEntity.ok().build();
    }

    @PutMapping("/me/password")
    public ResponseEntity<Void> changePassword(
            HttpServletRequest request,
            @RequestBody PasswordChangeRequest passwordRequest) {
        String token = request.getHeader("Authorization").substring(7);
        userService.changePassword(token, passwordRequest.getOldPassword(), passwordRequest.getNewPassword());
        return ResponseEntity.ok().build();
    }
}