// src/components/AuthRedirect.js

import React, { useEffect } from 'react';

const AuthRedirect = () => {
  useEffect(() => {
    // Force a page reload to ensure the session cookie is registered
    window.location.reload();
  }, []);

  return <div>Loading...</div>; // Optional: show a loading indicator while reloading
};

export default AuthRedirect;
