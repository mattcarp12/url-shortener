// src/App.tsx
import { useState } from 'react'
import './App.css' // Import our new vanilla CSS

function App() {
  const [originalUrl, setOriginalUrl] = useState('')
  const [customAlias, setCustomAlias] = useState('')
  const [shortUrl, setShortUrl] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState('')

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsLoading(true)
    setError('')
    setShortUrl('')

    try {
      const response = await fetch('/api/urls', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          original_url: originalUrl,
          custom_alias: customAlias || undefined,
        }),
      })

      if (!response.ok) {
        throw new Error('Failed to create short URL. Alias might be taken.')
      }

      const data = await response.json()
      setShortUrl(data.short_url)
      setOriginalUrl('')
      setCustomAlias('')
    } catch (err: any) {
      setError(err.message)
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="container">
      <h1 className="title">Bitly-Clone</h1>
      <p className="subtitle">Paste your long URL below to shorten it.</p>

      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label>
            Long URL <span className="required">*</span>
          </label>
          <input
            type="url"
            required
            placeholder="https://example.com/very-long-link"
            className="input-field"
            value={originalUrl}
            onChange={(e) => setOriginalUrl(e.target.value)}
          />
        </div>

        <div className="form-group">
          <label>Custom Alias (Optional)</label>
          <div className="alias-input-group">
            <span className="alias-prefix">localhost:5173/</span>
            <input
              type="text"
              placeholder="my-campaign"
              className="input-field"
              value={customAlias}
              onChange={(e) => setCustomAlias(e.target.value)}
            />
          </div>
        </div>

        <button type="submit" disabled={isLoading || !originalUrl} className="submit-btn">
          {isLoading ? 'Shortening...' : 'Shorten URL'}
        </button>
      </form>

      {error && <div className="error-message">{error}</div>}

      {shortUrl && (
        <div className="success-message">
          <p>Success! Your short URL is:</p>
          <a href={shortUrl} target="_blank" rel="noopener noreferrer" className="short-link">
            {shortUrl}
          </a>
        </div>
      )}
    </div>
  )
}

export default App