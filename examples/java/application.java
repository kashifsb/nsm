package com.nsm.example;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.boot.context.event.ApplicationReadyEvent;
import org.springframework.context.event.EventListener;

@SpringBootApplication
public class Application {

    public static void main(String[] args) {
        SpringApplication.run(Application.class, args);
    }

    @EventListener(ApplicationReadyEvent.class)
    public void onApplicationReady() {
        String port = System.getProperty("server.port", "{{.Port}}");
        String domain = "{{.Domain}}";
        String nsmEnabled = System.getenv("NSM_ENABLED");
        
        System.out.println();
        System.out.println("🚀 Java Spring Boot server started");
        System.out.println("🌐 Domain: " + domain);
        System.out.println("📡 NSM: " + ("true".equals(nsmEnabled) ? "Enabled" : "Disabled"));
        System.out.println("☕ Java: " + System.getProperty("java.version"));
        System.out.println("🍃 Spring Boot: Ready");
        System.out.println();
    }
}
