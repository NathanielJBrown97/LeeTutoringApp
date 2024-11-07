// src/components/SignIn.js

import React from 'react';
import { API_BASE_URL } from '../config';
import './SignIn.css'; // Import the CSS file for styling

const SignIn = () => {
  // Handler for Google Sign-In
  const handleGoogleLogin = () => {
    window.location.href = `${API_BASE_URL}/internal/googleauth/oauth`;
  };

  // Handler for Microsoft Sign-In
  const handleMicrosoftLogin = () => {
    window.location.href = `${API_BASE_URL}/internal/microsoftauth/oauth`;
  };

  // Handler for Disabled Buttons
  const handleDisabledClick = (provider) => {
    alert(`Sign in with ${provider} is coming soon!`);
  };

  return (
    <div className="signin-container">
      <h1>Welcome to Agora</h1>
      <p>Please select a platform to sign in:</p>
      <div className="button-container">
        <button className="signin-button google" onClick={handleGoogleLogin}>
          Sign in with Google
        </button>
        <button className="signin-button microsoft" onClick={handleMicrosoftLogin}>
          Sign in with Microsoft
        </button>
        <button
          className="signin-button facebook disabled"
          onClick={() => handleDisabledClick('Facebook')}
          disabled
        >
          Sign in with Facebook (Coming Soon)
        </button>
        <button
          className="signin-button yahoo disabled"
          onClick={() => handleDisabledClick('Yahoo')}
          disabled
        >
          Sign in with Yahoo (Coming Soon)
        </button>
        <button
          className="signin-button apple disabled"
          onClick={() => handleDisabledClick('Apple')}
          disabled
        >
          Sign in with Apple (Coming Soon)
        </button>
      </div>
    </div>
  );
};

export default SignIn;
