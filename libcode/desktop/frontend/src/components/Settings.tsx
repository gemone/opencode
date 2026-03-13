import React, { useState } from 'react';
import { Config } from '../types';
import './Settings.css';

interface SettingsProps {
  config: Config;
  onSave: (config: Config) => void;
  onClose: () => void;
}

export const Settings: React.FC<SettingsProps> = ({ config, onSave, onClose }) => {
  const [localConfig, setLocalConfig] = useState<Config>({ ...config });

  const handleChange = (key: keyof Config, value: any) => {
    setLocalConfig((prev) => ({ ...prev, [key]: value }));
  };

  const handleSave = () => {
    onSave(localConfig);
  };

  return (
    <div className="settings">
      <div className="settings-header">
        <h2>Settings</h2>
        <button className="settings-close-button" onClick={onClose}>
          <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
            <path d="M4.646 4.646a.5.5 0 0 1 .708 0L8 7.293l2.646-2.647a.5.5 0 0 1 .708.708L8.707 8l2.647 2.646a.5.5 0 0 1-.708.708L8 8.707l-2.646 2.647a.5.5 0 0 1-.708-.708L7.293 8 4.646 5.354a.5.5 0 0 1 0-.708z" />
          </svg>
        </button>
      </div>

      <div className="settings-content">
        <section className="settings-section">
          <h3>API Configuration</h3>
          <div className="settings-field">
            <label htmlFor="apiKey">API Key</label>
            <input
              id="apiKey"
              type="password"
              value={localConfig.apiKey || ''}
              onChange={(e) => handleChange('apiKey', e.target.value)}
              placeholder="Enter your API key"
            />
          </div>

          <div className="settings-field">
            <label htmlFor="model">Model</label>
            <input
              id="model"
              type="text"
              value={localConfig.model || ''}
              onChange={(e) => handleChange('model', e.target.value)}
              placeholder="e.g., claude-sonnet-4-6"
            />
          </div>

          <div className="settings-field">
            <label htmlFor="maxTokens">Max Tokens</label>
            <input
              id="maxTokens"
              type="number"
              value={localConfig.maxTokens || 4096}
              onChange={(e) => handleChange('maxTokens', parseInt(e.target.value))}
              min={1}
              max={100000}
            />
          </div>

          <div className="settings-field">
            <label htmlFor="temperature">Temperature: {localConfig.temperature?.toFixed(2) || 0.7}</label>
            <input
              id="temperature"
              type="range"
              value={localConfig.temperature || 0.7}
              onChange={(e) => handleChange('temperature', parseFloat(e.target.value))}
              min={0}
              max={2}
              step={0.1}
            />
          </div>
        </section>

        <section className="settings-section">
          <h3>General</h3>
          <div className="settings-field">
            <label htmlFor="theme">Theme</label>
            <select
              id="theme"
              value={localConfig.theme || 'dark'}
              onChange={(e) => handleChange('theme', e.target.value)}
            >
              <option value="dark">Dark</option>
              <option value="light">Light</option>
              <option value="auto">Auto</option>
            </select>
          </div>

          <div className="settings-field">
            <label htmlFor="logLevel">Log Level</label>
            <select
              id="logLevel"
              value={localConfig.logLevel || 'INFO'}
              onChange={(e) => handleChange('logLevel', e.target.value)}
            >
              <option value="DEBUG">Debug</option>
              <option value="INFO">Info</option>
              <option value="WARN">Warning</option>
              <option value="ERROR">Error</option>
            </select>
          </div>

          <div className="settings-field-checkbox">
            <input
              id="autoSave"
              type="checkbox"
              checked={localConfig.autoSave || false}
              onChange={(e) => handleChange('autoSave', e.target.checked)}
            />
            <label htmlFor="autoSave">Auto-save conversations</label>
          </div>
        </section>

        <section className="settings-section">
          <h3>Storage</h3>
          <div className="settings-field">
            <label htmlFor="storagePath">Storage Path</label>
            <input
              id="storagePath"
              type="text"
              value={localConfig.storagePath || ''}
              onChange={(e) => handleChange('storagePath', e.target.value)}
              placeholder="Default: ~/.libcode"
            />
          </div>
        </section>
      </div>

      <div className="settings-footer">
        <button className="settings-button settings-button-secondary" onClick={onClose}>
          Cancel
        </button>
        <button className="settings-button settings-button-primary" onClick={handleSave}>
          Save Settings
        </button>
      </div>
    </div>
  );
};
