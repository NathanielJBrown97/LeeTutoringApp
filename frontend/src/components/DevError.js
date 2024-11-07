// src/components/DevError.js

import React, { useContext } from 'react';
import { AuthContext } from '../contexts/AuthContext';
import './DevError.css'; // Optional: Create CSS for styling

const DevError = () => {
  const authState = useContext(AuthContext);

  return (
    <div className="deverror-container">
      <h1>Authentication Failed</h1>
      <p>We encountered an issue while trying to authenticate your session.</p>
      {authState.error && (
        <div className="error-details">
          <h2>Error Details:</h2>
          <p>{authState.error}</p>
        </div>
      )}
      <p>Please try signing in again.</p>
      <a href="/" className="retry-link">Go to Sign-In Page</a>
    </div>
  );
};

export default DevError;
