// src/components/Login.js

import React from 'react';

const Login = () => {
  const handleGoogleLogin = () => {
    window.location.href = '/internal/googleauth/oauth';
  };

  const handleMicrosoftLogin = () => {
    window.location.href = '/internal/microsoftauth/oauth';
  };

  return (
    <div style={styles.container}>
      <h1 style={styles.title}>Welcome to Agora</h1>
      <div style={styles.buttonContainer}>
        <button style={styles.button} onClick={handleGoogleLogin}>
          Sign in with Google
        </button>
        <button style={styles.button} onClick={handleMicrosoftLogin}>
          Sign in with Microsoft
        </button>
      </div>
    </div>
  );
};

const styles = {
  container: {
    textAlign: 'center',
    marginTop: '20%',
  },
  title: {
    fontSize: '2em',
  },
  buttonContainer: {
    marginTop: '2em',
  },
  button: {
    margin: '0 1em',
    padding: '1em 2em',
    fontSize: '1em',
  },
};

export default Login;
