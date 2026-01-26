# Frontend Setup

## Quick Start

```bash
# Install dependencies
npm install

# Create environment file
cp .env.example .env

# Edit .env with your backend URL
# REACT_APP_API_URL=http://localhost:8080/api

# Start development server
npm start
```

The app will open at http://localhost:3000

## Available Scripts

### `npm start`
Runs the app in development mode at http://localhost:3000

### `npm build`
Builds the app for production to the `build` folder

### `npm test`
Launches the test runner in interactive watch mode

## Environment Variables

Create a `.env` file in the frontend directory:

```env
REACT_APP_API_URL=http://localhost:8080/api
```

For production, create `.env.production`:

```env
REACT_APP_API_URL=https://your-backend-url.com/api
```

## Project Structure

```
src/
├── components/
│   ├── ShareMode.js       # File upload and note creation
│   ├── ShareMode.css
│   ├── AccessMode.js      # Content retrieval by ID
│   └── AccessMode.css
├── App.js                 # Main app component
├── App.css
├── config.js              # API configuration
└── index.js               # Entry point
```

## Features

### Share Mode
- Drag and drop file uploads
- Browse and select files
- Create text notes with titles
- Real-time upload progress
- Copy share ID or URL to clipboard

### Access Mode
- Search by share ID
- View notes inline
- Download files
- URL-based sharing (add ?id=XXX to URL)

## Deployment

### Vercel (Recommended)

1. Install Vercel CLI:
```bash
npm install -g vercel
```

2. Deploy:
```bash
vercel --prod
```

3. Set environment variables in Vercel dashboard:
   - Go to Settings → Environment Variables
   - Add `REACT_APP_API_URL` with your backend URL

### Build for Static Hosting

```bash
npm run build
```

Upload the `build/` directory to:
- Netlify
- GitHub Pages
- AWS S3 + CloudFront
- Any static host

### Docker

```bash
docker build -t konbi-frontend .
docker run -p 80:80 konbi-frontend
```

## Customization

### Colors and Styling

Main colors are defined in CSS files:
- Primary: `#667eea` (purple-blue)
- Secondary: `#764ba2` (purple)

Edit in `App.css` and component CSS files.

### API Configuration

Update [config.js](src/config.js) to change API URL logic:

```javascript
const API_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080/api';
```

### Share URL Format

To customize share URL format, edit the `getShareUrl()` function in [ShareMode.js](src/components/ShareMode.js):

```javascript
const getShareUrl = () => {
  return `${window.location.origin}?id=${result.id}`;
};
```

## Browser Support

- Chrome (latest)
- Firefox (latest)
- Safari (latest)
- Edge (latest)

## Performance

- Code splitting enabled
- Lazy loading for better performance
- Optimized production builds
- Asset caching

## Troubleshooting

### API Connection Issues

Check that:
1. Backend is running
2. `.env` file has correct API URL
3. CORS is configured in backend
4. No firewall blocking requests

### Build Errors

Clear cache and reinstall:
```bash
rm -rf node_modules package-lock.json
npm install
npm run build
```

### Upload Issues

- Check file size (max 50MB)
- Verify file type is allowed
- Check browser console for errors
