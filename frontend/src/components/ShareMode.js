import React, { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { HiCloudUpload, HiDocumentText, HiCheckCircle, HiXCircle, HiClipboard, HiFolder, HiX } from 'react-icons/hi';
import axios from 'axios';
import API_URL from '../config';
import './ShareMode.css';

function ShareMode() {
  const [activeTab, setActiveTab] = useState('file'); // 'file' or 'note'
  const [file, setFile] = useState(null);
  const [noteTitle, setNoteTitle] = useState('');
  const [noteContent, setNoteContent] = useState('');
  const [uploading, setUploading] = useState(false);
  const [result, setResult] = useState(null);
  const [error, setError] = useState(null);
  const [dragActive, setDragActive] = useState(false);
  const [uploadProgress, setUploadProgress] = useState(0);

  const handleDrag = (e) => {
    e.preventDefault();
    e.stopPropagation();
    if (e.type === 'dragenter' || e.type === 'dragover') {
      setDragActive(true);
    } else if (e.type === 'dragleave') {
      setDragActive(false);
    }
  };

  const handleDrop = (e) => {
    e.preventDefault();
    e.stopPropagation();
    setDragActive(false);

    if (e.dataTransfer.files && e.dataTransfer.files[0]) {
      setFile(e.dataTransfer.files[0]);
      setError(null);
    }
  };

  const handleFileChange = (e) => {
    if (e.target.files && e.target.files[0]) {
      setFile(e.target.files[0]);
      setError(null);
    }
  };

  const handleFileUpload = async () => {
    if (!file) {
      setError('Please select a file');
      return;
    }

    if (file.size > 50 * 1024 * 1024) {
      setError('File size exceeds 50MB limit');
      return;
    }

    setUploading(true);
    setError(null);
    setResult(null);
    setUploadProgress(0);

    const formData = new FormData();
    formData.append('file', file);

    try {
      const response = await axios.post(`${API_URL}/upload`, formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
        onUploadProgress: (progressEvent) => {
          const progress = Math.round((progressEvent.loaded * 100) / progressEvent.total);
          setUploadProgress(progress);
        },
      });

      setResult({
        type: 'file',
        id: response.data.id,
        filename: response.data.filename,
        expiresAt: response.data.expiresAt,
      });
      setFile(null);
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to upload file');
    } finally {
      setUploading(false);
      setUploadProgress(0);
    }
  };

  const handleNoteSubmit = async () => {
    if (!noteContent.trim()) {
      setError('Please enter note content');
      return;
    }

    setUploading(true);
    setError(null);
    setResult(null);

    try {
      const response = await axios.post(`${API_URL}/note`, {
        title: noteTitle || 'Untitled Note',
        content: noteContent,
      });

      setResult({
        type: 'note',
        id: response.data.id,
        title: response.data.title,
        expiresAt: response.data.expiresAt,
      });
      setNoteTitle('');
      setNoteContent('');
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to create note');
    } finally {
      setUploading(false);
    }
  };

  const copyToClipboard = (text) => {
    navigator.clipboard.writeText(text);
    alert('Copied to clipboard!');
  };

  const getShareUrl = () => {
    return `${window.location.origin}?id=${result.id}`;
  };

  return (
    <div className="share-mode">
      <div className="tabs">
        <motion.button
          className={activeTab === 'file' ? 'active' : ''}
          onClick={() => setActiveTab('file')}
          whileHover={{ scale: 1.02 }}
          whileTap={{ scale: 0.98 }}
        >
          <HiCloudUpload size={20} />
          <span>Upload File</span>
        </motion.button>
        <motion.button
          className={activeTab === 'note' ? 'active' : ''}
          onClick={() => setActiveTab('note')}
          whileHover={{ scale: 1.02 }}
          whileTap={{ scale: 0.98 }}
        >
          <HiDocumentText size={20} />
          <span>Create Note</span>
        </motion.button>
      </div>

      <AnimatePresence mode="wait">
        {activeTab === 'file' ? (
          <motion.div
            key="file"
            className="file-upload"
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -10 }}
            transition={{ duration: 0.3 }}
          >
            <motion.div
              className={`drop-zone ${dragActive ? 'drag-active' : ''} ${file ? 'has-file' : ''}`}
              onDragEnter={handleDrag}
              onDragLeave={handleDrag}
              onDragOver={handleDrag}
              onDrop={handleDrop}
              animate={dragActive ? { scale: 1.02 } : { scale: 1 }}
              transition={{ duration: 0.2 }}
            >
              {file ? (
                <motion.div
                  className="file-info"
                  initial={{ opacity: 0, scale: 0.9 }}
                  animate={{ opacity: 1, scale: 1 }}
                  transition={{ duration: 0.3 }}
                >
                  <HiCheckCircle size={48} className="file-icon success" />
                  <p className="file-name">{file.name}</p>
                  <p className="file-size">{(file.size / 1024 / 1024).toFixed(2)} MB</p>
                  <motion.button
                    className="remove-file"
                    onClick={() => setFile(null)}
                    whileHover={{ scale: 1.05 }}
                    whileTap={{ scale: 0.95 }}
                  >
                    <HiX size={16} />
                    Remove
                  </motion.button>
                </motion.div>
              ) : (
                <>
                  <motion.div
                    animate={{ y: [0, -10, 0] }}
                    transition={{ duration: 2, repeat: Infinity, ease: "easeInOut" }}
                  >
                    <HiCloudUpload size={64} className="drop-icon" />
                  </motion.div>
                  <p>Drag and drop your file here</p>
                  <p className="or">or</p>
                  <motion.label
                    className="file-input-label"
                    whileHover={{ scale: 1.05 }}
                    whileTap={{ scale: 0.95 }}
                  >
                    <HiFolder size={20} />
                    Browse Files
                    <input
                      type="file"
                      onChange={handleFileChange}
                      style={{ display: 'none' }}
                    />
                  </motion.label>
                </>
              )}
            </motion.div>

            {uploading && uploadProgress > 0 && (
              <motion.div
                className="progress-bar"
                initial={{ opacity: 0, scale: 0.95 }}
                animate={{ opacity: 1, scale: 1 }}
              >
                <motion.div
                  className="progress-fill"
                  initial={{ width: 0 }}
                  animate={{ width: `${uploadProgress}%` }}
                  transition={{ duration: 0.3 }}
                />
                <span className="progress-text">{uploadProgress}%</span>
              </motion.div>
            )}

            <motion.button
              className="submit-btn"
              onClick={handleFileUpload}
              disabled={!file || uploading}
              whileHover={!file || uploading ? {} : { scale: 1.02 }}
              whileTap={!file || uploading ? {} : { scale: 0.98 }}
            >
              {uploading ? 'Uploading...' : 'Upload File'}
            </motion.button>
          </motion.div>
        ) : (
          <motion.div
            key="note"
            className="note-create"
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -10 }}
            transition={{ duration: 0.3 }}
          >
            <input
              type="text"
              placeholder="Note Title (optional)"
              value={noteTitle}
              onChange={(e) => setNoteTitle(e.target.value)}
              className="note-title-input"
            />
            <textarea
              placeholder="Enter your note content here..."
              value={noteContent}
              onChange={(e) => setNoteContent(e.target.value)}
              className="note-content-input"
              rows="12"
            />
            <motion.button
              className="submit-btn"
              onClick={handleNoteSubmit}
              disabled={!noteContent.trim() || uploading}
              whileHover={!noteContent.trim() || uploading ? {} : { scale: 1.02 }}
              whileTap={!noteContent.trim() || uploading ? {} : { scale: 0.98 }}
            >
              {uploading ? 'Creating...' : 'Create Note'}
            </motion.button>
          </motion.div>
        )}
      </AnimatePresence>

      <AnimatePresence>
        {error && (
          <motion.div
            className="error-message"
            initial={{ opacity: 0, y: -10 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -10 }}
          >
            <HiXCircle size={20} />
            {error}
          </motion.div>
        )}
      </AnimatePresence>

      <AnimatePresence>
        {result && (
          <motion.div
            className="result-box"
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            exit={{ opacity: 0, scale: 0.95 }}
            transition={{ duration: 0.3 }}
          >
            <h3>
              <HiCheckCircle size={28} />
              Success!
            </h3>
            <div className="result-content">
              <div className="id-display">
                <label>Share ID:</label>
                <div className="id-value">
                  <code>{result.id}</code>
                  <motion.button
                    onClick={() => copyToClipboard(result.id)}
                    whileHover={{ scale: 1.05 }}
                    whileTap={{ scale: 0.95 }}
                  >
                    <HiClipboard size={18} />
                    Copy ID
                  </motion.button>
                </div>
              </div>
              <div className="url-display">
                <label>Share URL:</label>
                <div className="url-value">
                  <input type="text" value={getShareUrl()} readOnly />
                  <motion.button
                    onClick={() => copyToClipboard(getShareUrl())}
                    whileHover={{ scale: 1.05 }}
                    whileTap={{ scale: 0.95 }}
                  >
                    <HiClipboard size={18} />
                    Copy URL
                  </motion.button>
                </div>
              </div>
              <p className="expires-text">
                Expires: {new Date(result.expiresAt).toLocaleString()}
              </p>
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}

export default ShareMode;
