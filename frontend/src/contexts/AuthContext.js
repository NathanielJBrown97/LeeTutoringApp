// src/contexts/AuthContext.js

import React, { createContext, useEffect, useState } from 'react';
import { API_BASE_URL } from '../config';

export const AuthContext = createContext();

export const AuthProvider = ({ children }) => {
  const [authState, setAuthState] = useState({
    authenticated: false,
    user: null,
    loading: true,
    error: null,
  });

  useEffect(() => {
    const fetchAuthStatus = () => {
      fetch(`${API_BASE_URL}/api/auth/status`, {
        method: 'GET',
        credentials: 'include', // Important to include cookies
      })
        .then(async (response) => {
          if (!response.ok) {
            const errorText = await response.text();
            throw new Error(`Error ${response.status}: ${errorText}`);
          }
          return response.json();
        })
        .then((data) => {
          console.log("Auth status fetched:", data);
          if (data.authenticated) {
            setAuthState({
              authenticated: true,
              user: { id: data.userID, email: data.email },
              loading: false,
              error: null,
            });
          } else {
            setAuthState({
              authenticated: false,
              user: null,
              loading: false,
              error: 'User is not authenticated.',
            });
          }
        })
        .catch((error) => {
          console.error('Error fetching auth status:', error);
          setAuthState({
            authenticated: false,
            user: null,
            loading: false,
            error: error.message,
          });
        });
    };

    // Delay the initial check to ensure the session cookie is set
    setTimeout(fetchAuthStatus, 500); // Adjust the delay as needed

  }, []);

  return (
    <AuthContext.Provider value={authState}>
      {children}
    </AuthContext.Provider>
  );
};
