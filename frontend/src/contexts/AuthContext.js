// src/contexts/AuthContext.js

import React, { createContext, useEffect, useState } from 'react';
import { jwtDecode } from 'jwt-decode';
import { API_BASE_URL } from '../config';

export const AuthContext = createContext();

export const AuthProvider = ({ children }) => {
  const [token, setToken] = useState(localStorage.getItem('authToken'));
  const [authState, setAuthState] = useState({
    authenticated: false,
    user: null,
    loading: true,
    error: null,
  });

  useEffect(() => {
    console.log('Token from state:', token);

    if (token) {
      try {
        // Decode the token to get user information including role
        const decoded = jwtDecode(token);
        console.log('Decoded token:', decoded);

        // Check if the token is expired
        const currentTime = Date.now() / 1000; // in seconds
        if (decoded.exp < currentTime) {
          console.log('Token has expired');
          localStorage.removeItem('authToken');
          setToken(null);
          setAuthState({
            authenticated: false,
            user: null,
            loading: false,
            error: 'Session has expired. Please sign in again.',
          });
        } else {
          console.log('Token is valid');
          // Store user information including role in the state
          setAuthState({
            authenticated: true,
            user: {
              id: decoded.user_id,
              email: decoded.email,
              role: decoded.role, // Ensure your token includes this property
              associatedStudents: decoded.associated_students || [],
            },
            loading: false,
            error: null,
          });
        }
      } catch (error) {
        console.error('Error decoding token:', error);
        localStorage.removeItem('authToken');
        setToken(null);
        setAuthState({
          authenticated: false,
          user: null,
          loading: false,
          error: 'Invalid token. Please sign in again.',
        });
      }
    } else {
      console.log('No token found');
      setAuthState({
        authenticated: false,
        user: null,
        loading: false,
        error: null,
      });
    }
  }, [token]);

  // Function to update the token state
  const updateToken = (newToken) => {
    setToken(newToken);
  };

  const contextValue = {
    ...authState,
    updateToken,
  };

  return (
    <AuthContext.Provider value={contextValue}>
      {children}
    </AuthContext.Provider>
  );
};
