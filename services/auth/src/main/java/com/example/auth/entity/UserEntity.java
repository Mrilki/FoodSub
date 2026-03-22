package com.example.auth.entity;

import com.example.auth.enums.UserRole;
import jakarta.persistence.*;
import lombok.*;
import org.hibernate.annotations.CreationTimestamp;
import org.hibernate.annotations.UpdateTimestamp;

import java.time.LocalDateTime;
import java.util.UUID;

@Setter
@Getter
@NoArgsConstructor
@AllArgsConstructor
@Builder

@Entity
@Table(name = "users")
public class UserEntity {

    @Id
    @GeneratedValue(strategy = GenerationType.UUID)
    private UUID id;

//    public UUID getId() {
//        return this.id;
//    }

    @Column(nullable = false, unique = true)
    private String email;

//    public String getEmail() {
//        return this.email;
//    }

    @Column(name = "password_hash", nullable = false)
    private String passwordHash;

//    public String getPasswordHash() {
//        return this.passwordHash;
//    }

    @Enumerated(EnumType.STRING)
    @Column(nullable = false)
    private UserRole role;

//    public UserRole getRole() {
//        return this.role;
//    }

    @Column(name = "is_active", nullable = false)
    private boolean active;

//    public boolean isActive() {
//        return active;
//    }

    @CreationTimestamp
    @Column(name = "created_at", updatable = false)
    private LocalDateTime createdAt;

    @UpdateTimestamp
    @Column(name = "updated_at")
    private LocalDateTime updatedAt;

    // Связь 1:1 с профилем
    @OneToOne(mappedBy = "user", cascade = CascadeType.ALL, fetch = FetchType.LAZY)
    private UserProfileEntity profile;
}