package com.nsm.example.model;

import com.fasterxml.jackson.annotation.JsonProperty;

import java.time.Instant;

public class Task {
    private String id;
    private String title;
    private String description;
    private String priority;
    private String status;
    
    @JsonProperty("created_at")
    private Instant createdAt;

    // Builder pattern
    public static Builder builder() {
        return new Builder();
    }

    public static class Builder {
        private Task task = new Task();

        public Builder id(String id) {
            task.id = id;
            return this;
        }

        public Builder title(String title) {
            task.title = title;
            return this;
        }

        public Builder description(String description) {
            task.description = description;
            return this;
        }

        public Builder priority(String priority) {
            task.priority = priority;
            return this;
        }

        public Builder status(String status) {
            task.status = status;
            return this;
        }

        public Builder createdAt(Instant createdAt) {
            task.createdAt = createdAt;
            return this;
        }

        public Task build() {
            return task;
        }
    }

    // Getters and setters
    public String getId() { return id; }
    public void setId(String id) { this.id = id; }

    public String getTitle() { return title; }
    public void setTitle(String title) { this.title = title; }

    public String getDescription() { return description; }
    public void setDescription(String description) { this.description = description; }

    public String getPriority() { return priority; }
    public void setPriority(String priority) { this.priority = priority; }

    public String getStatus() { return status; }
    public void setStatus(String status) { this.status = status; }

    public Instant getCreatedAt() { return createdAt; }
    public void setCreatedAt(Instant createdAt) { this.createdAt = createdAt; }
}
