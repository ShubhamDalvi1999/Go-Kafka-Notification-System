import React, { useState, useEffect } from 'react';
import { NotificationBadge } from './components/NotificationBadge';
import { NotificationCenter } from './components/NotificationCenter';
import { EngagementDashboard } from './components/EngagementDashboard';
import { NotificationTester } from './components/NotificationTester';
import { NotificationSyncManager } from './utils/NotificationSyncManager';
import './App.css';

// Mock user ID for demo purposes
const DEMO_USER_ID = '0241733d-3384-483a-9fd9-5d373e5a2fe6';

function App() {
  const [isNotificationCenterOpen, setIsNotificationCenterOpen] = useState(false);
  const [isEngagementDashboardOpen, setIsEngagementDashboardOpen] = useState(false);
  const [isNotificationTesterOpen, setIsNotificationTesterOpen] = useState(false);
  const [syncManager] = useState(() => new NotificationSyncManager('web-app-1'));

  // Real-time disabled: WebSocket removed

  // Handle notification badge click
  const handleNotificationBadgeClick = () => {
    setIsNotificationCenterOpen(true);
  };

  // Handle notification center close
  const handleNotificationCenterClose = () => {
    setIsNotificationCenterOpen(false);
  };

  // Handle engagement dashboard toggle
  const handleEngagementDashboardToggle = () => {
    setIsEngagementDashboardOpen(!isEngagementDashboardOpen);
  };

  // Sync status indicator
  const syncStatus = syncManager.getSyncStatus();

  return (
    <div className="app">
      {/* Header */}
      <header className="app-header">
        <div className="header-left">
          <h1>🎓 Learning Platform</h1>
        </div>
        
        <div className="header-right">
          {/* Real-time disabled */}
          
          {/* Sync Status */}
          <div className="sync-status">
            <span className="sync-indicator">
              {syncStatus.pendingOperations > 0 ? '🔄' : '✅'}
            </span>
            <span className="sync-count">
              {syncStatus.pendingOperations} pending
            </span>
          </div>
          
          {/* Notification Badge */}
          <div className="notification-badge-container" onClick={handleNotificationBadgeClick}>
            <NotificationBadge 
              userId={DEMO_USER_ID} 
              className="header-notification-badge"
            />
          </div>
          
          {/* Notification Tester Toggle */}
          <button 
            onClick={() => setIsNotificationTesterOpen(!isNotificationTesterOpen)}
            className="tester-toggle-btn"
          >
            🧪 Test Notifications
          </button>
          
          {/* User Menu */}
          <div className="user-menu">
            <button className="user-menu-button">
              👤 User
            </button>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="app-main">
        <div className="content-header">
          <h2>Welcome to Your Learning Journey!</h2>
          <p>Track your progress, maintain streaks, and stay motivated with smart notifications.</p>
        </div>

        {/* Quick Actions */}
        <div className="quick-actions">
          <button 
            className="action-button primary"
            onClick={handleEngagementDashboardToggle}
          >
            {isEngagementDashboardOpen ? 'Hide' : 'Show'} Engagement Dashboard
          </button>
          
          <button className="action-button secondary">
            Start Practice Session
          </button>
          
          <button className="action-button secondary">
            View Courses
          </button>
        </div>

        {/* Engagement Dashboard */}
        {isEngagementDashboardOpen && (
          <div className="dashboard-container">
            <EngagementDashboard userId={DEMO_USER_ID} />
          </div>
        )}

        {/* Notification Tester */}
        {isNotificationTesterOpen && (
          <div className="tester-container">
            <NotificationTester />
          </div>
        )}

        {/* Demo Content */}
        <div className="demo-content">
          <div className="demo-section">
            <h3>🚀 Getting Started</h3>
            <p>
              This is a demo of the enhanced notification system with real-time updates, 
              engagement tracking, and smart scheduling. The system includes:
            </p>
            <ul>
              <li>✅ Real-time notifications via WebSocket</li>
              <li>✅ Engagement dashboard with streak tracking</li>
              <li>✅ Notification preferences management</li>
              <li>✅ Multi-device synchronization</li>
              <li>✅ Automated notification scheduling</li>
              <li>✅ Race condition handling</li>
            </ul>
          </div>

          <div className="demo-section">
            <h3>🔔 Notification Types</h3>
            <p>The system supports various notification types:</p>
            <div className="notification-types-grid">
              <div className="notification-type-card">
                <span className="type-icon">📅</span>
                <h4>Daily Reminders</h4>
                <p>Keep your practice streak alive</p>
              </div>
              
              <div className="notification-type-card">
                <span className="type-icon">🔥</span>
                <h4>Streak Reminders</h4>
                <p>Don't break your streak!</p>
              </div>
              
              <div className="notification-type-card">
                <span className="type-icon">🏆</span>
                <h4>Achievement Unlocks</h4>
                <p>Celebrate your progress</p>
              </div>
              
              <div className="notification-type-card">
                <span className="type-icon">📊</span>
                <h4>Weekly Recaps</h4>
                <p>Review your weekly progress</p>
              </div>
            </div>
          </div>

          <div className="demo-section">
            <h3>⚡ Real-time Features</h3>
            <p>Experience the power of real-time updates:</p>
            <ul>
              <li><strong>WebSocket Connection:</strong> Live notification delivery</li>
              <li><strong>Multi-device Sync:</strong> Consistent state across devices</li>
              <li><strong>Conflict Resolution:</strong> Smart handling of race conditions</li>
              <li><strong>Optimistic Updates:</strong> Immediate UI feedback</li>
            </ul>
          </div>
        </div>
      </main>

      {/* Notification Center */}
      <NotificationCenter
        userId={DEMO_USER_ID}
        isOpen={isNotificationCenterOpen}
        onClose={handleNotificationCenterClose}
      />

      {/* Footer */}
      <footer className="app-footer">
        <div className="footer-content">
          <p>&copy; 2024 Learning Platform. Enhanced notification system demo.</p>
          <div className="footer-links">
            <button onClick={() => syncManager.forceSync()}>
              Force Sync
            </button>
            <button onClick={() => console.log(syncManager.getSyncStatus())}>
              Debug Sync
            </button>
          </div>
        </div>
      </footer>

    </div>
  );
}

export default App;
