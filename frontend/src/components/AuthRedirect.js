// src/components/AuthRedirect.js

import React, { useEffect, useContext } from 'react';
import { useNavigate } from 'react-router-dom';
import { AuthContext } from '../contexts/AuthContext';

const AuthRedirect = () => {
  const navigate = useNavigate();
  const { updateToken } = useContext(AuthContext);

  useEffect(() => {
    // Extract the token from the URL fragment
    const token = window.location.hash.substr(1); // Remove the '#' at the beginning
    console.log('Extracted token:', token);

    if (token) {
      // Store the token in localStorage
      localStorage.setItem('authToken', token);

      // Update the token in AuthContext
      updateToken(token);

      // Remove the token from the URL to clean up
      window.history.replaceState({}, document.title, window.location.pathname);

      // Redirect to the dashboard
      navigate('/parentdashboard');
      console.log('Navigated to /parentdashboard');
    } else {
      // If no token is found, redirect to the sign-in page
      navigate('/');
      console.log('Navigated to /');
    }
  }, [navigate, updateToken]);

  return <div>Loading...</div>; // Optional: show a loading indicator while redirecting
};

export default AuthRedirect;
