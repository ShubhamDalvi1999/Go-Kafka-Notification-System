-- Initial database schema for Kafka Notification System
-- Migration: 001_initial_schema.sql

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create enum types
CREATE TYPE notification_type AS ENUM (
    'daily_reminder',
    'streak_reminder', 
    'last_chance_alert',
    'achievement_unlock',
    'xp_goal_reminder',
    'league_update',
    'we_miss_you',
    'event_notification',
    'new_course',
    'practice_needed',
    'weekly_recap'
);

CREATE TYPE notification_channel AS ENUM (
    'in_app',
    'push',
    'email',
    'sms'
);

CREATE TYPE delivery_status AS ENUM (
    'queued',
    'sent',
    'delivered',
    'failed',
    'suppressed',
    'read'
);

CREATE TYPE priority_level AS ENUM (
    'low',
    'medium',
    'high',
    'urgent'
);

-- Create users table
CREATE TABLE users (
    user_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    total_xp INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create user_profiles table
CREATE TABLE user_profiles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    full_name VARCHAR(255),
    avatar_url TEXT,
    bio TEXT,
    username VARCHAR(100) UNIQUE,
    location VARCHAR(255),
    website TEXT,
    skills TEXT[], -- Array of skills
    role VARCHAR(100) DEFAULT 'user',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create notification_templates table
CREATE TABLE notification_templates (
    id BIGSERIAL PRIMARY KEY,
    type notification_type NOT NULL,
    channel notification_channel NOT NULL,
    title VARCHAR(255),
    body TEXT NOT NULL,
    locale VARCHAR(10) DEFAULT 'en',
    priority priority_level DEFAULT 'medium',
    is_active BOOLEAN DEFAULT true,
    version INTEGER DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create notifications table
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    type notification_type NOT NULL,
    channel notification_channel NOT NULL,
    priority priority_level DEFAULT 'medium',
    template_id BIGINT REFERENCES notification_templates(id),
    title VARCHAR(255),
    message TEXT NOT NULL,
    metadata JSONB DEFAULT '{}',
    dedupe_key VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    scheduled_for TIMESTAMP WITH TIME ZONE,
    sent_at TIMESTAMP WITH TIME ZONE,
    delivered_at TIMESTAMP WITH TIME ZONE,
    read_at TIMESTAMP WITH TIME ZONE,
    status delivery_status DEFAULT 'queued'
);

-- Create user_notification_preferences table
CREATE TABLE user_notification_preferences (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    type notification_type NOT NULL,
    channel notification_channel NOT NULL,
    enabled BOOLEAN DEFAULT true,
    quiet_hours_start VARCHAR(5), -- Format: "HH:MM"
    quiet_hours_end VARCHAR(5),   -- Format: "HH:MM"
    max_per_day INTEGER,
    last_sent_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, type, channel)
);

-- Create notification_delivery_attempts table
CREATE TABLE notification_delivery_attempts (
    id BIGSERIAL PRIMARY KEY,
    notification_id UUID NOT NULL REFERENCES notifications(id) ON DELETE CASCADE,
    attempt_no INTEGER NOT NULL,
    status delivery_status NOT NULL,
    error_code VARCHAR(100),
    error_message TEXT,
    provider_message_id VARCHAR(255),
    latency_ms INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create outbox_notifications table (for Kafka)
CREATE TABLE outbox_notifications (
    id BIGSERIAL PRIMARY KEY,
    notification_id UUID NOT NULL REFERENCES notifications(id) ON DELETE CASCADE,
    topic VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL,
    published BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    published_at TIMESTAMP WITH TIME ZONE
);

-- Create user_engagement_streaks table
CREATE TABLE user_engagement_streaks (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    streak_type VARCHAR(100) NOT NULL,
    current_streak INTEGER DEFAULT 0,
    longest_streak INTEGER DEFAULT 0,
    last_activity_date DATE,
    streak_start_date DATE,
    total_activities INTEGER DEFAULT 0,
    timezone VARCHAR(100) DEFAULT 'UTC',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, streak_type)
);

-- Create indexes for better performance
CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_type ON notifications(type);
CREATE INDEX idx_notifications_status ON notifications(status);
CREATE INDEX idx_notifications_scheduled_for ON notifications(scheduled_for);
CREATE INDEX idx_notifications_created_at ON notifications(created_at);

CREATE INDEX idx_user_preferences_user_id ON user_notification_preferences(user_id);
CREATE INDEX idx_user_preferences_type_channel ON user_notification_preferences(type, channel);

CREATE INDEX idx_outbox_notifications_published ON outbox_notifications(published);
CREATE INDEX idx_outbox_notifications_topic ON outbox_notifications(topic);

CREATE INDEX idx_engagement_streaks_user_id ON user_engagement_streaks(user_id);
CREATE INDEX idx_engagement_streaks_streak_type ON user_engagement_streaks(streak_type);

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Add triggers for updated_at
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_profiles_updated_at BEFORE UPDATE ON user_profiles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_notification_preferences_updated_at BEFORE UPDATE ON user_notification_preferences
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_engagement_streaks_updated_at BEFORE UPDATE ON user_engagement_streaks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert some sample data
INSERT INTO users (name, email, total_xp) VALUES 
    ('John Doe', 'john@example.com', 150),
    ('Jane Smith', 'jane@example.com', 300),
    ('Bob Johnson', 'bob@example.com', 75)
ON CONFLICT (email) DO NOTHING;

INSERT INTO notification_templates (type, channel, title, body, priority) VALUES
    ('daily_reminder', 'in_app', 'Daily Practice Reminder', 'Time for your daily practice! Keep your streak going.', 'medium'),
    ('streak_reminder', 'push', 'Streak Alert', 'Don''t break your streak! Practice now to keep it alive.', 'high'),
    ('achievement_unlock', 'email', 'Achievement Unlocked!', 'Congratulations! You''ve unlocked a new achievement.', 'medium');
