import React, { useState, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { 
  HiSearch, 
  HiDocument, 
  HiDownload, 
  HiClipboard,
  HiCheckCircle,
  HiXCircle,
  HiInbox
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

  useEffect(() => {
    // check if id is in url
    const params = new URLSearchParams(window.location.search);
    const urlId = params.get('id');
    if (urlId) {
      setId(urlId);
      handleFetch(urlId);
    }
  }, []);

  const handleFetch = async (contentId = id) => {
    if (!contentId.trim()) {
      setError('Please enter an ID');
      return;
    }

    setLoading(true);
    setError(null);
    setContent(null);

    try {
      const response = await axios.get(`${API_URL}/content/${contentId}`);
      setContent(response.data);
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to retrieve content');
    } finally {
      setLoading(false);
    }
  };

  const handleDownload = () => {
    // create download link
    const downloadUrl = `${API_URL}/content/${id}/download`;
    window.location.href = downloadUrl;
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
        <p className="subtitle">Enter the share ID to access files or notes</p>

        <div className="search-bar">
          <input
            type="text"
            placeholder="Enter share ID (e.g., AbC123Xy)"
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

      <AnimatePresence mode="wait">
        {error && (
          <motion.div 
            className="error-message"
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            exit={{ opacity: 0, scale: 0.95 }}
            transition={{ duration: 0.3 }}
          >
            <HiXCircle size={20} />
            {error}
          </motion.div>
        )}

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
