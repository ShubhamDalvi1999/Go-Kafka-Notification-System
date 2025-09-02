# Frontend Application

React-based frontend for the Real-Time Notification System with real-time updates, engagement tracking, and responsive design.

## 🚀 Quick Start

### Prerequisites
- Node.js 18+
- npm or yarn

### Installation

```bash
# Install dependencies
npm install

# Start development server
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview
```

## 🏗️ Architecture

The frontend follows modern React patterns with a component-based architecture:

```
├── src/
│   ├── components/           # React components
│   │   ├── NotificationBadge.tsx      # Unread count badge
│   │   ├── NotificationCenter.tsx     # Main notification center
│   │   ├── NotificationItem.tsx       # Individual notification
│   │   ├── NotificationPreferences.tsx # User preferences
│   │   └── EngagementDashboard.tsx    # Engagement tracking
│   ├── hooks/                # Custom React hooks
│   │   └── useWebSocket.ts   # WebSocket management
│   ├── utils/                # Utility functions
│   │   └── NotificationSyncManager.ts # Multi-device sync
│   ├── App.tsx               # Main application component
│   └── main.tsx              # Application entry point
├── public/                   # Static assets
└── package.json              # Dependencies and scripts
```

## 🎯 Features

### Real-Time Notifications
- **Live Updates**: WebSocket integration for instant notifications
- **Notification Badge**: Animated unread count indicator
- **Priority Colors**: Visual priority indicators (urgent, high, medium, low)
- **Type Icons**: Distinct icons for different notification types

### Notification Management
- **Mark as Read**: Individual and bulk read operations
- **Filtering**: By type, priority, and status
- **Search**: Full-text search across notifications
- **Pagination**: Efficient loading of large notification lists

### User Preferences
- **Channel Control**: Enable/disable specific notification channels
- **Type Preferences**: Granular control over notification types
- **Quiet Hours**: Configure do-not-disturb periods
- **Rate Limiting**: Set maximum notifications per day

### Engagement Dashboard
- **Streak Tracking**: Visual streak counters and progress
- **Progress Charts**: Time-based progress visualization
- **Achievements**: Milestone tracking and celebrations
- **Motivational Tips**: Contextual encouragement messages

### Multi-Device Support
- **State Synchronization**: Consistent state across devices
- **Conflict Resolution**: Smart handling of concurrent updates
- **Offline Support**: Queue operations for when connection returns
- **Real-Time Sync**: Automatic synchronization every 30 seconds

## 🛠️ Technology Stack

- **React 18**: Modern UI framework with hooks
- **TypeScript**: Type-safe development
- **Vite**: Fast build tool and dev server
- **Tailwind CSS**: Utility-first styling
- **WebSocket**: Real-time communication
- **Zustand**: Lightweight state management

## 📱 Components

### NotificationBadge
Displays unread notification count with animations and priority-based styling.

```tsx
<NotificationBadge 
  count={unreadCount} 
  onClick={openNotificationCenter}
/>
```

### NotificationCenter
Main notification management interface with tabs, filtering, and search.

```tsx
<NotificationCenter 
  notifications={notifications}
  onMarkAsRead={handleMarkAsRead}
  onUpdatePreferences={handleUpdatePreferences}
/>
```

### EngagementDashboard
User engagement tracking with streaks, progress, and achievements.

```tsx
<EngagementDashboard 
  user={currentUser}
  streaks={userStreaks}
  achievements={userAchievements}
/>
```

## 🔌 Hooks

### useWebSocket
Manages WebSocket connection with automatic reconnection and heartbeat monitoring.

```tsx
const { 
  isConnected, 
  sendMessage, 
  lastMessage 
} = useWebSocket('ws://localhost:8081/ws');
```

## 🔄 State Management

The application uses a combination of:
- **Local State**: Component-level state with `useState`
- **Context**: Theme and user preferences
- **WebSocket**: Real-time updates and synchronization
- **Local Storage**: Persistent user preferences

## 🎨 Styling

Built with **Tailwind CSS** for:
- **Responsive Design**: Mobile-first approach
- **Dark Mode**: Automatic theme switching
- **Animations**: Smooth transitions and micro-interactions
- **Accessibility**: High contrast and keyboard navigation

## 🧪 Testing

```bash
# Run tests
npm test

# Run tests with coverage
npm run test:coverage

# Run tests in watch mode
npm run test:watch
```

## 🚀 Development

### Development Server
```bash
npm run dev
```
Access the application at `http://localhost:5173`

### Building for Production
```bash
npm run build
```
Creates optimized production bundle in `dist/` folder

### Code Quality
```bash
# Lint code
npm run lint

# Format code
npm run format

# Type checking
npm run type-check
```

## 🔧 Configuration

### Environment Variables
Create a `.env.local` file for local development:

```bash
VITE_API_BASE_URL=http://localhost:8082
VITE_WS_URL=ws://localhost:8081/ws
VITE_APP_NAME=Notification System
```

### Build Configuration
The build is configured in `vite.config.ts` with:
- TypeScript support
- CSS preprocessing
- Asset optimization
- Development server configuration

## 📱 Responsive Design

The application is designed for:
- **Mobile**: 320px - 768px
- **Tablet**: 768px - 1024px
- **Desktop**: 1024px+

Key responsive features:
- Collapsible navigation
- Touch-friendly interactions
- Adaptive layouts
- Mobile-optimized notifications

## 🔒 Security

- **Input Validation**: Client-side validation for all forms
- **XSS Prevention**: Sanitized HTML rendering
- **CSRF Protection**: Secure API communication
- **Content Security Policy**: Restricted resource loading

## 📊 Performance

- **Code Splitting**: Lazy-loaded components
- **Bundle Optimization**: Tree shaking and minification
- **Image Optimization**: WebP format and lazy loading
- **Caching**: Service worker for offline support

## 🚀 Deployment

### Static Hosting
```bash
npm run build
# Deploy dist/ folder to your hosting provider
```

### Docker Deployment
```dockerfile
FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production
COPY . .
RUN npm run build
EXPOSE 80
CMD ["npm", "run", "preview", "--", "--host", "0.0.0.0", "--port", "80"]
```

## 🔍 Troubleshooting

### Common Issues

1. **WebSocket Connection Failed**
   - Check backend service is running
   - Verify WebSocket URL in configuration
   - Check firewall settings

2. **Build Errors**
   - Clear node_modules and reinstall
   - Check TypeScript configuration
   - Verify all dependencies are installed

3. **Runtime Errors**
   - Check browser console for errors
   - Verify API endpoints are accessible
   - Check environment variables

### Debug Mode

Enable debug logging in browser console:
```javascript
localStorage.setItem('debug', 'true')
```

## 📚 Documentation

- **[Component API](./docs/components.md)**: Detailed component documentation
- **[State Management](./docs/state.md)**: State management patterns
- **[Styling Guide](./docs/styling.md)**: CSS and design system
- **[Testing Guide](./docs/testing.md)**: Testing strategies and examples

## 🤝 Contributing

1. Follow React best practices
2. Use TypeScript for all new code
3. Add tests for new components
4. Follow the established component patterns
5. Update documentation for new features

## 📄 License

This project is licensed under the MIT License.
