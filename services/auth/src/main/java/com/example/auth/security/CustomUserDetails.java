package com.example.auth.security;


import com.example.auth.entity.UserEntity;
import lombok.AllArgsConstructor;
import lombok.Data;
import org.springframework.security.core.GrantedAuthority;
import org.springframework.security.core.authority.SimpleGrantedAuthority;
import org.springframework.security.core.userdetails.UserDetails;

import java.util.Collection;
import java.util.UUID;
import java.util.stream.Collectors;

@AllArgsConstructor
public class CustomUserDetails implements UserDetails {

    private UUID id;
    private String email;
    private String password;
    private Collection<? extends GrantedAuthority> authorities;
    private boolean isActive;

    public static CustomUserDetails fromUserEntity(UserEntity user) {
        return new CustomUserDetails(
                user.getId(),
                user.getEmail(),
                user.getPasswordHash(),
                user.getRole() != null ?
                        java.util.List.of(new SimpleGrantedAuthority("ROLE_" + user.getRole().name())) :
                        java.util.List.of(),
                user.isActive()
        );
    }

    @Override
    public String getUsername() {
        return email; // Используем email как логин
    }

    @Override
    public boolean isAccountNonLocked() {
        return isActive;
    }

    @Override
    public String getPassword() {
        return password;
    }


    public String getIdAsString() {
        return id != null ? id.toString() : null;
    }
    @Override
    public Collection<? extends GrantedAuthority> getAuthorities() {
        return authorities;
    }
}
