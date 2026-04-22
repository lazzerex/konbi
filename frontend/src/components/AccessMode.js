import React, { useState, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import {
  HiSearch,
  HiDocument,
  HiDownload,
  HiClipboard,
  HiCheckCircle,
  HiXCircle,
  HiInbox,
  HiFolder,
  HiX
} from 'react-icons/hi';
import axios from 'axios';
import API_URL from '../config';
import './AccessMode.css';

function AccessMode() {
  const [id, setId] = useState('');
  const [loading, setLoading] = useState(false);
  const [content, setContent] = useState(null);
  const [error, setError] = useState(null);
  const [copied, setCopied] = useState(false);

  const handleFetch = async (input = id) => {
    const lookupValue = input.trim();
    if (!lookupValue) {
      setError('Please enter an ID');
      return;
    }

    setLoading(true);
    setError(null);
    setContent(null);

    try {
      const response = await axios.get(`${API_URL}/content/${lookupValue}`);
      setContent(response.data);
    } catch (err) {
      const status = err.response?.status;
      if (status === 429) {
        setError('Too many requests. Please wait a moment and try again.');
      } else {
        setError(err.response?.data?.error || 'Failed to retrieve content');
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const urlId = params.get('id');
    if (urlId) {
      setId(urlId);
      handleFetch(urlId);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const handleDownload = () => {
    const downloadUrl = `${API_URL}/content/${content.id}/download`;
    window.location.href = downloadUrl;
  };

  const handleBundleDownload = () => {
    const zipUrl = `${API_URL}/content/${content.id}/zip`;
    window.location.href = zipUrl;
  };

  const copyContent = () => {
    if (content && content.type === 'note') {
      navigator.clipboard.writeText(content.content);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  return (
    <div className="access-mode">
      <motion.div 
        className="search-section"
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5 }}
      >
        <div className="search-header">
          <HiSearch size={28} className="search-header-icon" />
          <h2>Retrieve Content</h2>
        </div>
        <p className="subtitle">Enter the share ID or short code to access files or notes</p>

        <div className="search-bar">
          <input
            type="text"
            placeholder="Enter share ID or code (e.g., AbC123Xy)"
            value={id}
            onChange={(e) => setId(e.target.value)}
            onKeyPress={(e) => e.key === 'Enter' && handleFetch()}
          />
          <motion.button 
            onClick={() => handleFetch()} 
            disabled={loading || !id.trim()}
            whileHover={{ scale: 1.02 }}
            whileTap={{ scale: 0.98 }}
          >
            {loading ? (
              <>
                <motion.div
                  animate={{ rotate: 360 }}
                  transition={{ duration: 1, repeat: Infinity, ease: "linear" }}
                  style={{ display: 'inline-block' }}
                >
                  <HiSearch size={18} />
                </motion.div>
                Loading...
              </>
            ) : (
              <>
                <HiSearch size={18} />
                Access
              </>
            )}
          </motion.button>
        </div>
      </motion.div>

      <AnimatePresence>
        {error && (
          <motion.div
            className="error-modal-overlay"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 0.2 }}
            onClick={() => setError(null)}
          >
            <motion.div
              className="error-modal"
              initial={{ opacity: 0, scale: 0.9, y: -20 }}
              animate={{ opacity: 1, scale: 1, y: 0 }}
              exit={{ opacity: 0, scale: 0.9, y: -20 }}
              transition={{ duration: 0.25 }}
              onClick={(e) => e.stopPropagation()}
            >
              <div className="error-modal-header">
                <HiXCircle size={22} className="error-modal-icon" />
                <span>Error</span>
                <button className="error-modal-close" onClick={() => setError(null)}>
                  <HiX size={18} />
                </button>
              </div>
              <p className="error-modal-message">{error}</p>
              <button className="error-modal-btn" onClick={() => setError(null)}>
                Dismiss
              </button>
            </motion.div>
          </motion.div>
        )}
      </AnimatePresence>

      <AnimatePresence mode="wait">

        {content && content.type === 'note' && (
          <motion.div 
            className="content-display note-display"
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -20 }}
            transition={{ duration: 0.4 }}
          >
            <div className="content-header">
              <div className="header-left">
                <HiDocument size={24} className="content-icon" />
                <h3>{content.title || 'Untitled Note'}</h3>
              </div>
              <motion.button 
                className="copy-btn" 
                onClick={copyContent}
                whileHover={{ scale: 1.05 }}
                whileTap={{ scale: 0.95 }}
              >
                {copied ? (
                  <>
                    <HiCheckCircle size={18} />
                    Copied!
                  </>
                ) : (
                  <>
                    <HiClipboard size={18} />
                    Copy
                  </>
                )}
              </motion.button>
            </div>
            <div className="note-content">
              <pre>{content.content}</pre>
            </div>
          </motion.div>
        )}

        {content && content.type === 'file' && (
          <motion.div 
            className="content-display file-display"
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -20 }}
            transition={{ duration: 0.4 }}
          >
            <motion.div 
              className="file-icon-large"
              initial={{ scale: 0 }}
              animate={{ scale: 1 }}
              transition={{ delay: 0.2, type: "spring", stiffness: 200 }}
            >
              <HiDocument size={64} />
            </motion.div>
            <h3>{content.filename}</h3>
            <p className="file-size">
              Size: {(content.size / 1024 / 1024).toFixed(2)} MB
            </p>
            <motion.button 
              className="download-btn" 
              onClick={handleDownload}
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
            >
              <HiDownload size={20} />
              Download File
            </motion.button>
          </motion.div>
        )}

        {content && content.type === 'bundle' && (
          <motion.div
            className="content-display file-display bundle-display"
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -20 }}
            transition={{ duration: 0.4 }}
          >
            <motion.div
              className="file-icon-large"
              initial={{ scale: 0 }}
              animate={{ scale: 1 }}
              transition={{ delay: 0.2, type: "spring", stiffness: 200 }}
            >
              <HiFolder size={64} />
            </motion.div>
            <h3>Bundle Ready</h3>
            <div className="bundle-meta">
              <span className="bundle-pill">{content.fileCount || 0} files</span>
              {content.id && <span className="bundle-pill bundle-pill-muted">ID: {content.id}</span>}
            </div>
            {Array.isArray(content.files) && content.files.length > 0 && (
              <div className="bundle-files">
                {content.files.map((f, idx) => (
                  <div key={f.id || `${f.filename}-${idx}`} className="bundle-file-row">
                    <span className="bundle-file-index">{idx + 1}</span>
                    <span className="bundle-file-name" title={f.filename || f.id}>{f.filename || f.id}</span>
                    {typeof f.size === 'number' && f.size > 0 && (
                      <span className="bundle-file-size">{(f.size / 1024 / 1024).toFixed(2)} MB</span>
                    )}
                  </div>
                ))}
              </div>
            )}
            <motion.button
              className="download-btn"
              onClick={handleBundleDownload}
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
            >
              <HiDownload size={20} />
              Download Bundle (ZIP)
            </motion.button>
          </motion.div>
        )}

        {!content && !error && !loading && (
          <motion.div 
            className="empty-state"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ duration: 0.5 }}
          >
            <motion.div 
              className="empty-icon"
              animate={{ 
                y: [0, -10, 0],
              }}
              transition={{ 
                duration: 2,
                repeat: Infinity,
                ease: "easeInOut"
              }}
            >
              <HiInbox size={64} />
            </motion.div>
            <p>Enter a share ID to view content</p>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}

export default AccessMode;
