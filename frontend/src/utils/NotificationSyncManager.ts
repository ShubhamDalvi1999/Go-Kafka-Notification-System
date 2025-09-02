import { Notification } from '../types/notification';

interface SyncOperation {
  id: string;
  type: 'mark_read' | 'mark_delivered' | 'create' | 'update';
  notificationId: string;
  timestamp: number;
  data?: any;
}

interface DeviceState {
  deviceId: string;
  lastSync: number;
  pendingOperations: SyncOperation[];
}

export class NotificationSyncManager {
  private deviceId: string;
  private pendingOperations: Map<string, SyncOperation> = new Map();
  private lastServerSync: number = 0;
  private syncInProgress: boolean = false;
  private syncInterval: NodeJS.Timeout | null = null;
  private conflictResolutionStrategy: 'server_wins' | 'client_wins' | 'timestamp_wins' = 'timestamp_wins';

  constructor(deviceId: string) {
    this.deviceId = deviceId;
    this.startPeriodicSync();
  }

  /**
   * Start periodic synchronization
   */
  private startPeriodicSync(): void {
    this.syncInterval = setInterval(() => {
      this.performSync();
    }, 30000); // Sync every 30 seconds
  }

  /**
   * Stop periodic synchronization
   */
  public stopPeriodicSync(): void {
    if (this.syncInterval) {
      clearInterval(this.syncInterval);
      this.syncInterval = null;
    }
  }

  /**
   * Add a pending operation to the queue
   */
  public addPendingOperation(operation: Omit<SyncOperation, 'id' | 'timestamp'>): void {
    const id = `${operation.type}_${operation.notificationId}_${Date.now()}`;
    const syncOperation: SyncOperation = {
      ...operation,
      id,
      timestamp: Date.now()
    };

    this.pendingOperations.set(id, syncOperation);
    console.log(`Added pending operation: ${id}`, syncOperation);
  }

  /**
   * Mark a notification as read
   */
  public markAsRead(notificationId: string): void {
    this.addPendingOperation({
      type: 'mark_read',
      notificationId,
      data: { readAt: new Date().toISOString() }
    });
  }

  /**
   * Mark a notification as delivered
   */
  public markAsDelivered(notificationId: string): void {
    this.addPendingOperation({
      type: 'mark_delivered',
      notificationId,
      data: { deliveredAt: new Date().toISOString() }
    });
  }

  /**
   * Create a new notification
   */
  public createNotification(notification: Notification): void {
    this.addPendingOperation({
      type: 'create',
      notificationId: notification.id,
      data: notification
    });
  }

  /**
   * Update an existing notification
   */
  public updateNotification(notificationId: string, updates: Partial<Notification>): void {
    this.addPendingOperation({
      type: 'update',
      notificationId,
      data: updates
    });
  }

  /**
   * Perform synchronization with the server
   */
  public async performSync(): Promise<void> {
    if (this.syncInProgress || this.pendingOperations.size === 0) {
      return;
    }

    this.syncInProgress = true;
    console.log(`Starting sync with ${this.pendingOperations.size} pending operations`);

    try {
      // Get all pending operations
      const operations = Array.from(this.pendingOperations.values());
      
      // Group operations by type for batch processing
      const operationsByType = this.groupOperationsByType(operations);
      
      // Process each type of operation
      for (const [type, typeOperations] of operationsByType) {
        await this.processOperationsByType(type, typeOperations);
      }

      // Update last sync timestamp
      this.lastServerSync = Date.now();
      
      console.log('Sync completed successfully');
    } catch (error) {
      console.error('Sync failed:', error);
      // Retry failed operations with exponential backoff
      this.retryFailedOperations();
    } finally {
      this.syncInProgress = false;
    }
  }

  /**
   * Group operations by type for efficient processing
   */
  private groupOperationsByType(operations: SyncOperation[]): Map<string, SyncOperation[]> {
    const grouped = new Map<string, SyncOperation[]>();
    
    for (const operation of operations) {
      if (!grouped.has(operation.type)) {
        grouped.set(operation.type, []);
      }
      grouped.get(operation.type)!.push(operation);
    }
    
    return grouped;
  }

  /**
   * Process operations of a specific type
   */
  private async processOperationsByType(type: string, operations: SyncOperation[]): Promise<void> {
    try {
      switch (type) {
        case 'mark_read':
          await this.processMarkAsReadOperations(operations);
          break;
        case 'mark_delivered':
          await this.processMarkAsDeliveredOperations(operations);
          break;
        case 'create':
          await this.processCreateOperations(operations);
          break;
        case 'update':
          await this.processUpdateOperations(operations);
          break;
        default:
          console.warn(`Unknown operation type: ${type}`);
      }
    } catch (error) {
      console.error(`Failed to process ${type} operations:`, error);
      throw error;
    }
  }

  /**
   * Process mark as read operations
   */
  private async processMarkAsReadOperations(operations: SyncOperation[]): Promise<void> {
    const notificationIds = operations.map(op => op.notificationId);
    
    // Batch API call to mark notifications as read
    await this.apiCall('markAsRead', { notificationIds });
    
    // Remove successful operations from pending queue
    operations.forEach(op => this.pendingOperations.delete(op.id));
  }

  /**
   * Process mark as delivered operations
   */
  private async processMarkAsDeliveredOperations(operations: SyncOperation[]): Promise<void> {
    const notificationIds = operations.map(op => op.notificationId);
    
    // Batch API call to mark notifications as delivered
    await this.apiCall('markAsDelivered', { notificationIds });
    
    // Remove successful operations from pending queue
    operations.forEach(op => this.pendingOperations.delete(op.id));
  }

  /**
   * Process create operations
   */
  private async processCreateOperations(operations: SyncOperation[]): Promise<void> {
    const notifications = operations.map(op => op.data);
    
    // Batch API call to create notifications
    await this.apiCall('createNotifications', { notifications });
    
    // Remove successful operations from pending queue
    operations.forEach(op => this.pendingOperations.delete(op.id));
  }

  /**
   * Process update operations
   */
  private async processUpdateOperations(operations: SyncOperation[]): Promise<void> {
    const updates = operations.map(op => ({
      id: op.notificationId,
      updates: op.data
    }));
    
    // Batch API call to update notifications
    await this.apiCall('updateNotifications', { updates });
    
    // Remove successful operations from pending queue
    operations.forEach(op => this.pendingOperations.delete(op.id));
  }

  /**
   * Make API call to the server
   */
  private async apiCall(endpoint: string, data: any): Promise<any> {
    // This would be replaced with actual API calls
    const response = await fetch(`/api/notifications/${endpoint}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(data),
    });

    if (!response.ok) {
      throw new Error(`API call failed: ${response.statusText}`);
    }

    return response.json();
  }

  /**
   * Retry failed operations with exponential backoff
   */
  private retryFailedOperations(): void {
    const operations = Array.from(this.pendingOperations.values());
    
    operations.forEach(operation => {
      // Add exponential backoff delay
      const delay = Math.min(1000 * Math.pow(2, Math.floor(Math.random() * 3)), 10000);
      
      setTimeout(() => {
        if (this.pendingOperations.has(operation.id)) {
          console.log(`Retrying operation: ${operation.id}`);
          this.performSync();
        }
      }, delay);
    });
  }

  /**
   * Handle conflicts between local and server state
   */
  public resolveConflict(localNotification: Notification, serverNotification: Notification): Notification {
    switch (this.conflictResolutionStrategy) {
      case 'server_wins':
        return serverNotification;
      
      case 'client_wins':
        return localNotification;
      
      case 'timestamp_wins':
        const localTimestamp = new Date(localNotification.updatedAt || localNotification.createdAt).getTime();
        const serverTimestamp = new Date(serverNotification.updatedAt || serverNotification.createdAt).getTime();
        return localTimestamp > serverTimestamp ? localNotification : serverNotification;
      
      default:
        return serverNotification;
    }
  }

  /**
   * Get current sync status
   */
  public getSyncStatus(): {
    pendingOperations: number;
    lastSync: number;
    isConnected: boolean;
    deviceId: string;
  } {
    return {
      pendingOperations: this.pendingOperations.size,
      lastSync: this.lastServerSync,
      isConnected: this.syncInProgress,
      deviceId: this.deviceId
    };
  }

  /**
   * Force immediate synchronization
   */
  public forceSync(): Promise<void> {
    return this.performSync();
  }

  /**
   * Clear all pending operations (useful for testing or reset)
   */
  public clearPendingOperations(): void {
    this.pendingOperations.clear();
    console.log('Cleared all pending operations');
  }

  /**
   * Set conflict resolution strategy
   */
  public setConflictResolutionStrategy(strategy: 'server_wins' | 'client_wins' | 'timestamp_wins'): void {
    this.conflictResolutionStrategy = strategy;
    console.log(`Conflict resolution strategy set to: ${strategy}`);
  }

  /**
   * Get pending operations for debugging
   */
  public getPendingOperations(): SyncOperation[] {
    return Array.from(this.pendingOperations.values());
  }
}
