// File: src/components/AuthRedirect.js
import React, { useEffect, useContext } from 'react';
import { useNavigate } from 'react-router-dom';
import { AuthContext } from '../contexts/AuthContext';
import { jwtDecode } from 'jwt-decode';

const AuthRedirect = () => {
  const navigate = useNavigate();
  const { updateToken } = useContext(AuthContext);

  useEffect(() => {
    // Extract the token from the URL fragment (removes the '#' at the beginning)
    const token = window.location.hash.substr(1);
    console.log('Extracted token:', token);

    if (token) {
      // Save token in localStorage and update our AuthContext state
      localStorage.setItem('authToken', token);
      updateToken(token);

      // Clean up the URL
      window.history.replaceState({}, document.title, window.location.pathname);

      // Decode the token to extract the user's role
      let decoded;
      try {
        decoded = jwtDecode(token);
        console.log('Decoded token:', decoded);
      } catch (error) {
        console.error('Error decoding token:', error);
      }

      // Based on the role, navigate to the appropriate dashboard
      if (decoded && decoded.role === 'tutor') {
        navigate('/tutordashboard');
        console.log('Navigated to /tutordashboard');
      } else if (decoded && decoded.role === 'student') {
        navigate('/studentdashboard');
        console.log('Navigated to /studentdashboard');
      } else {
        navigate('/parentdashboard');
        console.log('Navigated to /parentdashboard');
      }
    } else {
      // No token found: redirect to sign in
      navigate('/');
      console.log('Navigated to /');
    }
  }, [navigate, updateToken]);

  return <div>Loading...</div>;
};

export default AuthRedirect;
