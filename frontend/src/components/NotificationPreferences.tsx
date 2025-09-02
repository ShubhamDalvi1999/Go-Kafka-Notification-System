import React, { useState, useEffect } from 'react';
import { useNotificationPreferences } from '../hooks/useNotificationPreferences';
import './NotificationPreferences.css';

interface NotificationPreferencesProps {
  userId: string;
}

export const NotificationPreferences: React.FC<NotificationPreferencesProps> = ({
  userId
}) => {
  const {
    preferences,
    updatePreferences,
    isLoading,
    error,
    savePreferences
  } = useNotificationPreferences(userId);

  const [localPreferences, setLocalPreferences] = useState<typeof preferences>([]);
  const [hasChanges, setHasChanges] = useState(false);

  useEffect(() => {
    if (preferences.length > 0) {
      setLocalPreferences([...preferences]);
    }
  }, [preferences]);

  const handlePreferenceChange = (
    type: string,
    channel: string,
    field: string,
    value: any
  ) => {
    setLocalPreferences(prev => 
      prev.map(pref => {
        if (pref.type === type && pref.channel === channel) {
          return { ...pref, [field]: value };
        }
        return pref;
      })
    );
    setHasChanges(true);
  };

  const handleSave = async () => {
    try {
      await savePreferences(localPreferences);
      setHasChanges(false);
    } catch (err) {
      console.error('Failed to save preferences:', err);
    }
  };

  const handleReset = () => {
    setLocalPreferences([...preferences]);
    setHasChanges(false);
  };

  const getNotificationTypeLabel = (type: string) => {
    const labels: Record<string, string> = {
      daily_reminder: 'Daily Reminders',
      streak_reminder: 'Streak Reminders',
      last_chance_alert: 'Last Chance Alerts',
      achievement_unlock: 'Achievement Unlocks',
      xp_goal_reminder: 'XP Goal Reminders',
      league_update: 'League Updates',
      we_miss_you: 'Engagement Nudges',
      event_notification: 'Event Notifications',
      new_course: 'New Course Alerts',
      practice_needed: 'Practice Reminders',
      weekly_recap: 'Weekly Recaps'
    };
    return labels[type] || type.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase());
  };

  const getChannelLabel = (channel: string) => {
    const labels: Record<string, string> = {
      in_app: 'In-App',
      push: 'Push Notifications',
      email: 'Email',
      sms: 'SMS'
    };
    return labels[channel] || channel;
  };

  if (isLoading) {
    return (
      <div className="preferences-loading">
        <div className="spinner"></div>
        <p>Loading preferences...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="preferences-error">
        <p>Error loading preferences: {error}</p>
        <button onClick={() => window.location.reload()}>
          Retry
        </button>
      </div>
    );
  }

  return (
    <div className="notification-preferences">
      <div className="preferences-header">
        <h3>Notification Settings</h3>
        <p>Customize how and when you receive notifications</p>
      </div>

      {/* Global Settings */}
      <div className="preferences-section">
        <h4>Global Settings</h4>
        <div className="global-settings">
          <div className="setting-item">
            <label>
              <input
                type="checkbox"
                checked={localPreferences.every(p => p.enabled)}
                onChange={(e) => {
                  setLocalPreferences(prev => 
                    prev.map(p => ({ ...p, enabled: e.target.checked }))
                  );
                  setHasChanges(true);
                }}
              />
              Enable all notifications
            </label>
          </div>
        </div>
      </div>

      {/* Per-Type Preferences */}
      <div className="preferences-section">
        <h4>Notification Types</h4>
        <div className="type-preferences">
          {localPreferences.map((pref) => (
            <div key={`${pref.type}-${pref.channel}`} className="preference-item">
              <div className="preference-header">
                <div className="preference-info">
                  <h5>{getNotificationTypeLabel(pref.type)}</h5>
                  <span className="channel-label">{getChannelLabel(pref.channel)}</span>
                </div>
                
                <div className="preference-toggle">
                  <label className="toggle-switch">
                    <input
                      type="checkbox"
                      checked={pref.enabled}
                      onChange={(e) => handlePreferenceChange(
                        pref.type,
                        pref.channel,
                        'enabled',
                        e.target.checked
                      )}
                    />
                    <span className="toggle-slider"></span>
                  </label>
                </div>
              </div>

              {pref.enabled && (
                <div className="preference-options">
                  {/* Quiet Hours */}
                  <div className="option-group">
                    <label>Quiet Hours:</label>
                    <div className="time-inputs">
                      <input
                        type="time"
                        value={pref.quietHoursStart || ''}
                        onChange={(e) => handlePreferenceChange(
                          pref.type,
                          pref.channel,
                          'quietHoursStart',
                          e.target.value || null
                        )}
                        placeholder="Start time"
                      />
                      <span>to</span>
                      <input
                        type="time"
                        value={pref.quietHoursEnd || ''}
                        onChange={(e) => handlePreferenceChange(
                          pref.type,
                          pref.channel,
                          'quietHoursEnd',
                          e.target.value || null
                        )}
                        placeholder="End time"
                      />
                    </div>
                  </div>

                  {/* Rate Limiting */}
                  <div className="option-group">
                    <label>Max per day:</label>
                    <input
                      type="number"
                      min="1"
                      max="50"
                      value={pref.maxPerDay || ''}
                      onChange={(e) => handlePreferenceChange(
                        pref.type,
                        pref.channel,
                        'maxPerDay',
                        e.target.value ? parseInt(e.target.value) : null
                      )}
                      placeholder="Unlimited"
                    />
                  </div>
                </div>
              )}
            </div>
          ))}
        </div>
      </div>

      {/* Quick Presets */}
      <div className="preferences-section">
        <h4>Quick Presets</h4>
        <div className="preset-buttons">
          <button
            className="preset-button"
            onClick={() => {
              setLocalPreferences(prev => 
                prev.map(p => ({ ...p, enabled: true }))
              );
              setHasChanges(true);
            }}
          >
            Enable All
          </button>
          
          <button
            className="preset-button"
            onClick={() => {
              setLocalPreferences(prev => 
                prev.map(p => ({ ...p, enabled: false }))
              );
              setHasChanges(true);
            }}
          >
            Disable All
          </button>
          
          <button
            className="preset-button"
            onClick={() => {
              setLocalPreferences(prev => 
                prev.map(p => ({
                  ...p,
                  quietHoursStart: '22:00',
                  quietHoursEnd: '08:00'
                }))
              );
              setHasChanges(true);
            }}
          >
            Set Quiet Hours (10 PM - 8 AM)
          </button>
        </div>
      </div>

      {/* Action Buttons */}
      {hasChanges && (
        <div className="preferences-actions">
          <button
            className="action-button secondary"
            onClick={handleReset}
          >
            Reset Changes
          </button>
          
          <button
            className="action-button primary"
            onClick={handleSave}
          >
            Save Preferences
          </button>
        </div>
      )}

      {/* Help Text */}
      <div className="preferences-help">
        <h4>Need Help?</h4>
        <ul>
          <li><strong>Quiet Hours:</strong> Notifications won't be sent during these times</li>
          <li><strong>Rate Limiting:</strong> Maximum number of notifications per day for each type</li>
          <li><strong>Channel Preferences:</strong> Choose how you want to receive each notification type</li>
        </ul>
      </div>
    </div>
  );
};
