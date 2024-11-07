// src/App.js

import React, { useContext } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider, AuthContext } from './contexts/AuthContext';
import SignIn from './components/SignIn';
import ParentDashboard from './components/ParentDashboard';
import DevError from './components/DevError';
import AuthRedirect from './components/AuthRedirect';

function AppRoutes() {
  const authState = useContext(AuthContext);

  if (authState.loading) {
    return <div>Loading...</div>;
  }

  if (authState.authenticated) {
    return (
      <Routes>
        <Route path="/parentdashboard" element={<ParentDashboard />} />
        <Route path="/" element={<Navigate to="/parentdashboard" />} />
        <Route path="*" element={<Navigate to="/parentdashboard" />} />
      </Routes>
    );
  }

  if (authState.error) {
    return (
      <Routes>
        <Route path="/deverror" element={<DevError />} />
        <Route path="/" element={<SignIn />} />
        <Route path="*" element={<Navigate to="/deverror" />} />
      </Routes>
    );
  }

  return (
    <Routes>
      <Route path="/" element={<SignIn />} />
      <Route path="/parentdashboard" element={<Navigate to="/" />} />
      <Route path="/deverror" element={<DevError />} />
      <Route path="/auth-redirect" element={<AuthRedirect />} />
      <Route path="*" element={<Navigate to="/" />} />
    </Routes>
  );
}

function App() {
  return (
    <AuthProvider>
      <Router>
        <AppRoutes />
      </Router>
    </AuthProvider>
  );
}

export default App;
