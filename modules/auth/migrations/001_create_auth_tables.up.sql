-- Create auth_refresh_tokens table (Depends on users table)
CREATE TABLE auth_refresh_tokens (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    user_id BIGINT NOT NULL,
    token_hash TEXT NOT NULL,
    device_type VARCHAR(10) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked BOOLEAN DEFAULT FALSE,
    CONSTRAINT fk_refresh_tokens_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE INDEX idx_auth_refresh_tokens_deleted_at ON auth_refresh_tokens(deleted_at);
CREATE INDEX idx_auth_refresh_tokens_user_id ON auth_refresh_tokens(user_id);
CREATE INDEX idx_auth_refresh_tokens_token_hash ON auth_refresh_tokens(token_hash);

-- Create auth_login_activities table (Depends on users table)
CREATE TABLE auth_login_activities (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    user_id BIGINT NOT NULL,
    action VARCHAR(10) NOT NULL,
    "timestamp" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    ip_address VARCHAR(45),
    user_agent VARCHAR(255),
    device_type VARCHAR(10),
    CONSTRAINT fk_login_activities_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE INDEX idx_auth_login_activities_deleted_at ON auth_login_activities(deleted_at);
CREATE INDEX idx_auth_login_activities_user_id ON auth_login_activities(user_id);
CREATE INDEX idx_auth_login_activities_timestamp ON auth_login_activities("timestamp");

-- Create auth_otp_attempt_logs table (Depends on users table)
CREATE TABLE auth_otp_attempt_logs (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    user_id BIGINT NOT NULL,
    profile_id BIGINT NOT NULL,
    user_type VARCHAR(50) NOT NULL,
    "timestamp" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    was_valid BOOLEAN NOT NULL,
    ip_address VARCHAR(45),
    user_agent VARCHAR(255),
    device_type VARCHAR(10),
    CONSTRAINT fk_otp_logs_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE INDEX idx_auth_otp_attempt_logs_deleted_at ON auth_otp_attempt_logs(deleted_at);
CREATE INDEX idx_auth_otp_attempt_logs_user_id ON auth_otp_attempt_logs(user_id);
CREATE INDEX idx_auth_otp_attempt_logs_profile_id_user_type ON auth_otp_attempt_logs(profile_id, user_type);
CREATE INDEX idx_auth_otp_attempt_logs_timestamp ON auth_otp_attempt_logs("timestamp"); 