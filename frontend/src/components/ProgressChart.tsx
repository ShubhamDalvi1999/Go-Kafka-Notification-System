import React from 'react';

interface ProgressChartProps {
  weeklyProgress: number[];
}

export function ProgressChart({ weeklyProgress }: ProgressChartProps) {
  const days = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];
  const maxValue = Math.max(...weeklyProgress, 1);

  return (
    <div className="card">
      <h3 className="chart-title">📊 Weekly Progress</h3>
      <div className="chart-container">
        <div className="chart-bars">
          {weeklyProgress.map((value, index) => (
            <div key={index} className="chart-bar-container">
              <div 
                className="chart-bar" 
                style={{ 
                  height: `${(value / maxValue) * 100}%`,
                  backgroundColor: value > 0 ? '#3b82f6' : '#e5e7eb'
                }}
              >
                <span className="bar-value">{value}</span>
              </div>
              <span className="bar-label">{days[index]}</span>
            </div>
          ))}
        </div>
      </div>
      <div className="chart-summary">
        <p>Total this week: {weeklyProgress.reduce((sum, val) => sum + val, 0)} sessions</p>
      </div>
    </div>
  );
}
