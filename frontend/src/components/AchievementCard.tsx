import React from 'react';
import { Achievement } from '../hooks/useEngagementData';

interface AchievementCardProps {
  achievement: Achievement;
}

export function AchievementCard({ achievement }: AchievementCardProps) {
  return (
    <div className="achievement-card">
      <div className="achievement-icon">{achievement.icon}</div>
      <div className="achievement-content">
        <h4 className="achievement-name">{achievement.name}</h4>
        <p className="achievement-description">{achievement.description}</p>
        <span className="achievement-date">
          Unlocked: {new Date(achievement.unlockedAt).toLocaleDateString()}
        </span>
      </div>
    </div>
  );
}
