-- Migration: Add avatar_url column to profiles table
-- Date: 2026-02-04
-- Description: Adds avatar support for user profiles

-- Add avatar_url column with default value
ALTER TABLE profiles 
ADD COLUMN IF NOT EXISTS avatar_url TEXT NOT NULL DEFAULT '/avatars/default.png';

-- Update existing users to have the default avatar
UPDATE profiles 
SET avatar_url = '/avatars/default.png' 
WHERE avatar_url IS NULL OR avatar_url = '';

-- Create index for avatar_url lookups (optional, for performance)
CREATE INDEX IF NOT EXISTS idx_profiles_avatar_url ON profiles(avatar_url);

-- Add comment to document the column
COMMENT ON COLUMN profiles.avatar_url IS 'URL path to user avatar image. Points to predefined avatars in /avatars/ directory. Default: /avatars/default.png';
