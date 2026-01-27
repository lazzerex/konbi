import React, { useState, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { HiSun, HiMoon, HiUpload, HiSearch, HiLightningBolt, HiShieldCheck, HiClock, HiGlobe } from 'react-icons/hi';
import './App.css';
import ShareMode from './components/ShareMode';
import AccessMode from './components/AccessMode';

function App() {
  const [mode, setMode] = useState('share');
  const [theme, setTheme] = useState(() => {
    return localStorage.getItem('theme') || 'light';
  });

  useEffect(() => {
    document.documentElement.setAttribute('data-theme', theme);
    localStorage.setItem('theme', theme);
  }, [theme]);

  const toggleTheme = () => {
    setTheme(theme === 'light' ? 'dark' : 'light');
  };

  const features = [
    {
      icon: HiLightningBolt,
      title: 'Instant Sharing',
      description: 'Share files and notes in seconds. No sign-up required.',
      color: '#667eea'
    },
    {
      icon: HiShieldCheck,
      title: 'Secure & Private',
      description: 'End-to-end encryption keeps your content safe.',
      color: '#10b981'
    },
    {
      icon: HiClock,
      title: '7-Day Storage',
      description: 'Content automatically expires after 7 days.',
      color: '#f59e0b'
    },
    {
      icon: HiGlobe,
      title: 'Universal Access',
      description: 'Share with anyone, anywhere using a simple ID.',
      color: '#ec4899'
    }
  ];

  return (
    <div className="App">
      <motion.div 
        className="background-blur blur-1"
        animate={{ 
          scale: [1, 1.2, 1],
          rotate: [0, 90, 0]
        }}
        transition={{ 
          duration: 20, 
          repeat: Infinity,
          ease: "linear" 
        }}
      />
      <motion.div 
        className="background-blur blur-2"
        animate={{ 
          scale: [1, 1.3, 1],
          rotate: [180, 270, 180]
        }}
        transition={{ 
          duration: 25, 
          repeat: Infinity,
          ease: "linear" 
        }}
      />

      {/* Integrated Header with Actions */}
      <motion.header 
        className="App-header"
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.6, ease: "easeOut" }}
      >
        <div className="header-left">
          <motion.div 
            className="logo"
            initial={{ opacity: 0, scale: 0.9 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ duration: 0.5, delay: 0.2 }}
          >
            <img src="/konbi_logo.png" alt="konbi" className="logo-image" />
          </motion.div>
        </div>
        
        <div className="header-right">
          <motion.div 
            className="mode-toggle"
            initial={{ opacity: 0, x: 20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.5, delay: 0.3 }}
          >
            <motion.button
              className={mode === 'share' ? 'active' : ''}
              onClick={() => setMode('share')}
              whileHover={{ scale: 1.02 }}
              whileTap={{ scale: 0.98 }}
            >
              <HiUpload size={20} />
              <span>Share</span>
            </motion.button>
            <motion.button
              className={mode === 'access' ? 'active' : ''}
              onClick={() => setMode('access')}
              whileHover={{ scale: 1.02 }}
              whileTap={{ scale: 0.98 }}
            >
              <HiSearch size={20} />
              <span>Access</span>
            </motion.button>
          </motion.div>

          <motion.button 
            className="theme-toggle"
            onClick={toggleTheme}
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ duration: 0.5, delay: 0.4 }}
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
          >
            <AnimatePresence mode="wait">
              {theme === 'light' ? (
                <motion.div
                  key="moon"
                  initial={{ rotate: -180, opacity: 0 }}
                  animate={{ rotate: 0, opacity: 1 }}
                  exit={{ rotate: 180, opacity: 0 }}
                  transition={{ duration: 0.3 }}
                  style={{ display: 'flex', alignItems: 'center', gap: '8px' }}
                >
                  <HiMoon size={18} />
                  <span>Dark</span>
                </motion.div>
              ) : (
                <motion.div
                  key="sun"
                  initial={{ rotate: 180, opacity: 0 }}
                  animate={{ rotate: 0, opacity: 1 }}
                  exit={{ rotate: -180, opacity: 0 }}
                  transition={{ duration: 0.3 }}
                  style={{ display: 'flex', alignItems: 'center', gap: '8px' }}
                >
                  <HiSun size={18} />
                  <span>Light</span>
                </motion.div>
              )}
            </AnimatePresence>
          </motion.button>
        </div>
      </motion.header>

      {/* Asymmetric Grid Layout */}
      <div className="main-grid">
        {/* Left Panel - Feature Showcase */}
        <motion.div 
          className="info-panel"
          initial={{ opacity: 0, x: -30 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ duration: 0.7, delay: 0.2 }}
        >
          <motion.div 
            className="hero-section"
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6, delay: 0.3 }}
          >
            <h2 className="hero-title">Share Anything, Instantly</h2>
            <p className="hero-description">
              Drop your files or create notes. Get a shareable link in seconds. 
              No accounts, no hassle.
            </p>
          </motion.div>

          <div className="features-grid">
            {features.map((feature, index) => (
              <motion.div 
                key={index}
                className="feature-card"
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.5, delay: 0.4 + index * 0.1 }}
                whileHover={{ 
                  scale: 1.02,
                  boxShadow: '0 8px 30px rgba(0, 0, 0, 0.12)'
                }}
              >
                <div className="feature-icon" style={{ backgroundColor: `${feature.color}15` }}>
                  <feature.icon size={24} style={{ color: feature.color }} />
                </div>
                <div className="feature-content">
                  <h3>{feature.title}</h3>
                  <p>{feature.description}</p>
                </div>
              </motion.div>
            ))}
          </div>

          <motion.div 
            className="stats-section"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ duration: 0.6, delay: 0.9 }}
          >
            <div className="stat-item">
              <div className="stat-value">50MB</div>
              <div className="stat-label">Max File Size</div>
            </div>
            <div className="stat-divider"></div>
            <div className="stat-item">
              <div className="stat-value">7 Days</div>
              <div className="stat-label">Storage Time</div>
            </div>
            <div className="stat-divider"></div>
            <div className="stat-item">
              <div className="stat-value">E2E</div>
              <div className="stat-label">Encrypted</div>
            </div>
          </motion.div>
        </motion.div>

        {/* Right Panel - Upload/Access Area */}
        <AnimatePresence mode="wait">
          <motion.div 
            key={mode}
            className="content-panel"
            initial={{ opacity: 0, x: 30 }}
            animate={{ opacity: 1, x: 0 }}
            exit={{ opacity: 0, x: -30 }}
            transition={{ duration: 0.5 }}
          >
            {mode === 'share' ? <ShareMode /> : <AccessMode />}
          </motion.div>
        </AnimatePresence>
      </div>
    </div>
  );
}

export default App;
