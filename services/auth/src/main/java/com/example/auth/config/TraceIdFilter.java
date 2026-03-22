package com.example.auth.config;

import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.slf4j.MDC;
import org.springframework.stereotype.Component;
import org.springframework.web.filter.OncePerRequestFilter;

import java.io.IOException;
import java.util.UUID;

@Component
public class TraceIdFilter extends OncePerRequestFilter {

    private static final String TRACE_ID = "trace_id";

    @Override
    protected void doFilterInternal(HttpServletRequest request,
                                    HttpServletResponse response,
                                    FilterChain filterChain)
            throws ServletException, IOException {

        try {
            // Берем trace_id из заголовка или генерируем новый
            String traceId = request.getHeader("X-Trace-Id");
            if (traceId == null || traceId.isEmpty()) {
                traceId = UUID.randomUUID().toString();
            }

            // Кладем в MDC (контекст лога для этого потока)
            MDC.put(TRACE_ID, traceId);

            // Добавляем в ответ, чтобы клиент мог видеть его
            response.setHeader("X-Trace-Id", traceId);

            filterChain.doFilter(request, response);

        } finally {
            // Обязательно чистим MDC после запроса!
            MDC.clear();
        }
    }
}