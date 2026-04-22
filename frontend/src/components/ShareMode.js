import React, { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { HiCloudUpload, HiDocumentText, HiCheckCircle, HiXCircle, HiClipboard, HiFolder, HiX, HiQrcode } from 'react-icons/hi';
import { QRCodeSVG } from 'qrcode.react';
import axios from 'axios';
import API_URL from '../config';
import './ShareMode.css';

function ShareMode() {
  const [activeTab, setActiveTab] = useState('file');

  // file upload state
  const [files, setFiles] = useState([]);
  const [uploading, setUploading] = useState(false);
  const [uploadingIndex, setUploadingIndex] = useState(null);
  const [uploadTotal, setUploadTotal] = useState(0);
  const [uploadProgress, setUploadProgress] = useState(0);
  const [results, setResults] = useState([]);
  const [bundleResult, setBundleResult] = useState(null);
  const [qrVisible, setQrVisible] = useState(new Set());

  // note state
  const [noteTitle, setNoteTitle] = useState('');
  const [noteContent, setNoteContent] = useState('');
  const [noteResult, setNoteResult] = useState(null);

  // shared
  const [dragActive, setDragActive] = useState(false);
  const [error, setError] = useState(null);
  const [copyToast, setCopyToast] = useState(null);

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
    if (e.dataTransfer.files && e.dataTransfer.files.length > 0) {
      addFiles(Array.from(e.dataTransfer.files));
    }
  };

  const handleFileChange = (e) => {
    if (e.target.files && e.target.files.length > 0) {
      addFiles(Array.from(e.target.files));
    }
    e.target.value = '';
  };

  const addFiles = (incoming) => {
    setFiles(prev => {
      const existing = new Set(prev.map(f => f.name));
      return [...prev, ...incoming.filter(f => !existing.has(f.name))];
    });
    setError(null);
  };

  const removeFile = (index) => {
    setFiles(prev => prev.filter((_, i) => i !== index));
  };

  const handleFileUpload = async () => {
    if (files.length === 0) return;

    setUploading(true);
    setError(null);
    setResults([]);
    setBundleResult(null);
    setQrVisible(new Set());

    if (files.length === 1) {
      const file = files[0];
      if (file.size > 50 * 1024 * 1024) {
        setError(`${file.name}: exceeds 50MB limit`);
        setUploading(false);
        return;
      }

      setUploadingIndex(0);
      setUploadTotal(1);
      setUploadProgress(0);

      const formData = new FormData();
      formData.append('file', file);

      try {
        const response = await axios.post(`${API_URL}/upload`, formData, {
          headers: { 'Content-Type': 'multipart/form-data' },
          onUploadProgress: (e) => {
            setUploadProgress(Math.round((e.loaded * 100) / e.total));
          },
        });
        setResults([{
          id: response.data.id,
          code: response.data.code,
          filename: response.data.filename,
          expiresAt: response.data.expiresAt,
        }]);
      } catch (err) {
        setError(err.response?.data?.error || 'upload failed');
      }
    } else {
      // multiple files — upload as a bundle
      for (const file of files) {
        if (file.size > 50 * 1024 * 1024) {
          setError(`${file.name}: exceeds 50MB limit`);
          setUploading(false);
          return;
        }
      }

      setUploadProgress(0);
      const formData = new FormData();
      files.forEach(f => formData.append('files', f));

      try {
        const response = await axios.post(`${API_URL}/bundle`, formData, {
          headers: { 'Content-Type': 'multipart/form-data' },
          onUploadProgress: (e) => {
            setUploadProgress(Math.round((e.loaded * 100) / e.total));
          },
        });
        setBundleResult({
          id: response.data.id,
          code: response.data.code,
          fileCount: response.data.fileCount,
          expiresAt: response.data.expiresAt,
        });
      } catch (err) {
        setError(err.response?.data?.error || 'bundle upload failed');
      }
    }

    setFiles([]);
    setUploading(false);
    setUploadingIndex(null);
    setUploadProgress(0);
  };

  const handleNoteSubmit = async () => {
    if (!noteContent.trim()) {
      setError('Please enter note content');
      return;
    }

    setUploading(true);
    setError(null);
    setNoteResult(null);

    try {
      const response = await axios.post(`${API_URL}/note`, {
        title: noteTitle || 'Untitled Note',
        content: noteContent,
      });

      setNoteResult({
        id: response.data.id,
        code: response.data.code,
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

  const copyToClipboard = (text, label) => {
    navigator.clipboard.writeText(text);
    setCopyToast(`${label} copied!`);
    setTimeout(() => setCopyToast(null), 3000);
  };

  const getShareUrl = (id) => `${window.location.origin}?id=${id}`;

  const toggleQR = (id) => {
    setQrVisible(prev => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  };

  const formatSize = (bytes) => {
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    return `${(bytes / 1024 / 1024).toFixed(2)} MB`;
  };

  const getBundleZipUrl = (id) => `${API_URL}/content/${id}/zip`;

  const uploadButtonLabel = () => {
    if (uploading && files.length === 1) return `Uploading ${(uploadingIndex ?? 0) + 1} of ${uploadTotal}...`;
    if (uploading) return `Uploading bundle...`;
    if (files.length > 1) return `Upload Bundle (${files.length} files)`;
    return 'Upload File';
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
              className={`drop-zone ${dragActive ? 'drag-active' : ''} ${files.length > 0 ? 'has-file' : ''}`}
              onDragEnter={handleDrag}
              onDragLeave={handleDrag}
              onDragOver={handleDrag}
              onDrop={handleDrop}
              animate={dragActive ? { scale: 1.02 } : { scale: 1 }}
              transition={{ duration: 0.2 }}
            >
              {files.length > 0 ? (
                <motion.div
                  className="file-list"
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                >
                  <div className="file-list-header">
                    <span className="file-list-count">
                      <HiCheckCircle size={16} />
                      {files.length} file{files.length > 1 ? 's' : ''} selected
                    </span>
                    <motion.label
                      className="add-more-label"
                      whileHover={{ scale: 1.05 }}
                      whileTap={{ scale: 0.95 }}
                    >
                      <HiFolder size={14} />
                      Add more
                      <input
                        type="file"
                        multiple
                        onChange={handleFileChange}
                        style={{ display: 'none' }}
                      />
                    </motion.label>
                  </div>
                  <div className="file-list-items">
                    {files.map((f, i) => (
                      <motion.div
                        key={f.name}
                        className="file-list-item"
                        initial={{ opacity: 0, x: -8 }}
                        animate={{ opacity: 1, x: 0 }}
                        transition={{ delay: i * 0.04 }}
                      >
                        <HiCheckCircle size={16} className="file-icon success" />
                        <span className="file-list-name" title={f.name}>{f.name}</span>
                        <span className="file-list-size">{formatSize(f.size)}</span>
                        <motion.button
                          className="remove-file-btn"
                          onClick={() => removeFile(i)}
                          disabled={uploading}
                          whileHover={{ scale: 1.1 }}
                          whileTap={{ scale: 0.9 }}
                        >
                          <HiX size={13} />
                        </motion.button>
                      </motion.div>
                    ))}
                  </div>
                </motion.div>
              ) : (
                <>
                  <motion.div
                    animate={{ y: [0, -10, 0] }}
                    transition={{ duration: 2, repeat: Infinity, ease: 'easeInOut' }}
                  >
                    <HiCloudUpload size={64} className="drop-icon" />
                  </motion.div>
                  <p>Drag and drop your files here</p>
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
                      multiple
                      onChange={handleFileChange}
                      style={{ display: 'none' }}
                    />
                  </motion.label>
                </>
              )}
            </motion.div>

            {uploading && uploadingIndex !== null && (
              <motion.div
                className="upload-status"
                initial={{ opacity: 0, scale: 0.95 }}
                animate={{ opacity: 1, scale: 1 }}
              >
                <p className="upload-status-text">
                  Uploading file {uploadingIndex + 1} of {uploadTotal}
                </p>
                <div className="progress-bar">
                  <motion.div
                    className="progress-fill"
                    initial={{ width: 0 }}
                    animate={{ width: `${uploadProgress}%` }}
                    transition={{ duration: 0.3 }}
                  />
                  <span className="progress-text">{uploadProgress}%</span>
                </div>
              </motion.div>
            )}

            <motion.button
              className="submit-btn"
              onClick={handleFileUpload}
              disabled={files.length === 0 || uploading}
              whileHover={files.length === 0 || uploading ? {} : { scale: 1.02 }}
              whileTap={files.length === 0 || uploading ? {} : { scale: 0.98 }}
            >
              {uploadButtonLabel()}
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

      {/* multi-file upload results */}
      <AnimatePresence>
        {results.length > 0 && (
          <motion.div
            className="results-list"
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -10 }}
            transition={{ duration: 0.3 }}
          >
            <div className="results-list-header">
              <HiCheckCircle size={22} />
              {results.length} file{results.length > 1 ? 's' : ''} uploaded
              <motion.button
                className="close-result"
                onClick={() => { setResults([]); setBundleResult(null); setQrVisible(new Set()); }}
                whileHover={{ scale: 1.1 }}
                whileTap={{ scale: 0.9 }}
              >
                <HiX size={18} />
              </motion.button>
            </div>

            {results.length > 0 && results.map((r, i) => (
              <motion.div
                key={r.id}
                className="result-card"
                initial={{ opacity: 0, y: 8 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: i * 0.06 }}
              >
                <div className="result-card-filename">
                  <HiCheckCircle size={16} className="file-icon success" />
                  <span>{r.filename}</span>
                </div>

                <div className="result-card-row">
                  <span className="result-card-label">Share URL</span>
                  <div className="result-card-value">
                    <input type="text" value={getShareUrl(r.id)} readOnly />
                    <motion.button
                      onClick={() => copyToClipboard(getShareUrl(r.id), 'URL')}
                      whileHover={{ scale: 1.05 }}
                      whileTap={{ scale: 0.95 }}
                    >
                      <HiClipboard size={15} />
                    </motion.button>
                    <motion.button
                      className="qr-toggle-btn"
                      onClick={() => toggleQR(r.id)}
                      whileHover={{ scale: 1.05 }}
                      whileTap={{ scale: 0.95 }}
                      title="Toggle QR code"
                    >
                      <HiQrcode size={15} />
                    </motion.button>
                  </div>
                </div>

                <AnimatePresence>
                  {qrVisible.has(r.id) && (
                    <motion.div
                      className="result-card-qr"
                      initial={{ opacity: 0, height: 0 }}
                      animate={{ opacity: 1, height: 'auto' }}
                      exit={{ opacity: 0, height: 0 }}
                      transition={{ duration: 0.2 }}
                    >
                      <QRCodeSVG value={getShareUrl(r.id)} size={140} level="M" includeMargin={true} />
                    </motion.div>
                  )}
                </AnimatePresence>

                <span className="result-card-expires">
                  Expires {new Date(r.expiresAt).toLocaleString()}
                </span>
              </motion.div>
            ))}
          </motion.div>
        )}
      </AnimatePresence>

      {/* bundle result */}
      <AnimatePresence>
        {bundleResult && (
          <motion.div
            className="result-box"
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            exit={{ opacity: 0, scale: 0.95 }}
            transition={{ duration: 0.3 }}
          >
            <div className="result-header">
              <h3>
                <HiCheckCircle size={28} />
                Bundle uploaded! ({bundleResult.fileCount} files)
              </h3>
              <motion.button
                className="close-result"
                onClick={() => setBundleResult(null)}
                whileHover={{ scale: 1.1 }}
                whileTap={{ scale: 0.9 }}
              >
                <HiX size={20} />
              </motion.button>
            </div>
            <div className="result-content">
              <div className="url-display">
                <label>Share URL:</label>
                <div className="url-value">
                  <input type="text" value={getShareUrl(bundleResult.id)} readOnly />
                  <motion.button
                    onClick={() => copyToClipboard(getShareUrl(bundleResult.id), 'URL')}
                    whileHover={{ scale: 1.05 }}
                    whileTap={{ scale: 0.95 }}
                  >
                    <HiClipboard size={18} />
                    Copy URL
                  </motion.button>
                </div>
              </div>
              <div className="url-display">
                <label>Download All (ZIP):</label>
                <div className="url-value">
                  <input type="text" value={getBundleZipUrl(bundleResult.id)} readOnly />
                  <motion.button
                    onClick={() => copyToClipboard(getBundleZipUrl(bundleResult.id), 'ZIP URL')}
                    whileHover={{ scale: 1.05 }}
                    whileTap={{ scale: 0.95 }}
                  >
                    <HiClipboard size={18} />
                    Copy
                  </motion.button>
                </div>
              </div>
              <div className="qr-display">
                <label>Scan to share:</label>
                <div className="qr-code">
                  <QRCodeSVG
                    value={getShareUrl(bundleResult.id)}
                    size={160}
                    level="M"
                    includeMargin={true}
                  />
                </div>
              </div>
              <p className="expires-text">
                Expires: {new Date(bundleResult.expiresAt).toLocaleString()}
              </p>
            </div>
          </motion.div>
        )}
      </AnimatePresence>

      {/* note result */}
      <AnimatePresence>
        {noteResult && (
          <motion.div
            className="result-box"
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            exit={{ opacity: 0, scale: 0.95 }}
            transition={{ duration: 0.3 }}
          >
            <div className="result-header">
              <h3>
                <HiCheckCircle size={28} />
                Note created!
              </h3>
              <motion.button
                className="close-result"
                onClick={() => setNoteResult(null)}
                whileHover={{ scale: 1.1 }}
                whileTap={{ scale: 0.9 }}
              >
                <HiX size={20} />
              </motion.button>
            </div>
            <div className="result-content">
              <div className="id-display">
                <label>Share ID:</label>
                <div className="id-value">
                  <code>{noteResult.id}</code>
                  <motion.button
                    onClick={() => copyToClipboard(noteResult.id, 'ID')}
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
                  <input type="text" value={getShareUrl(noteResult.id)} readOnly />
                  <motion.button
                    onClick={() => copyToClipboard(getShareUrl(noteResult.id), 'URL')}
                    whileHover={{ scale: 1.05 }}
                    whileTap={{ scale: 0.95 }}
                  >
                    <HiClipboard size={18} />
                    Copy URL
                  </motion.button>
                </div>
              </div>
              <div className="qr-display">
                <label>Scan to share:</label>
                <div className="qr-code">
                  <QRCodeSVG
                    value={getShareUrl(noteResult.id)}
                    size={160}
                    level="M"
                    includeMargin={true}
                  />
                </div>
              </div>
              <p className="expires-text">
                Expires: {new Date(noteResult.expiresAt).toLocaleString()}
              </p>
            </div>
          </motion.div>
        )}
      </AnimatePresence>

      <AnimatePresence>
        {copyToast && (
          <motion.div
            className="copy-toast"
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: 20 }}
            transition={{ duration: 0.3 }}
          >
            <HiCheckCircle size={20} />
            {copyToast}
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}

export default ShareMode;
