import React, { useState, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { HiSun, HiMoon, HiUpload, HiSearch } from 'react-icons/hi';
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

  return (
    <div className="App">
      <motion.div 
        className="background-blur"
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

      <motion.button 
        className="theme-toggle"
        onClick={toggleTheme}
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

      <motion.header 
        className="App-header"
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.6, ease: "easeOut" }}
      >
        <motion.h1
          initial={{ opacity: 0, scale: 0.9 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ duration: 0.5, delay: 0.2 }}
        >
          konbi
        </motion.h1>
        <motion.p
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ duration: 0.5, delay: 0.4 }}
        >
          Share files and notes instantly
        </motion.p>
      </motion.header>

      <motion.div 
        className="mode-toggle"
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
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

      <AnimatePresence mode="wait">
        <motion.div 
          key={mode}
          className="content-container"
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          exit={{ opacity: 0, y: -20 }}
          transition={{ duration: 0.4 }}
        >
          {mode === 'share' ? <ShareMode /> : <AccessMode />}
        </motion.div>
      </AnimatePresence>

      <motion.footer 
        className="App-footer"
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ duration: 0.5, delay: 0.8 }}
      >
        <p>Expires in 7 days • Max 50MB • End-to-end encrypted</p>
      </motion.footer>
    </div>
  );
}

export default App;
