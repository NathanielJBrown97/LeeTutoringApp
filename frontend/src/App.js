// src/App.js

import React, { useContext, useEffect } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider, AuthContext } from './contexts/AuthContext';
import SignIn from './components/SignIn';
import ParentDashboard from './components/ParentDashboard';
import StudentIntake from './components/StudentIntake';
import BookingPage from './components/BookingPage'; // Import the new component
import AuthRedirect from './components/AuthRedirect';

function AppRoutes() {
  const authState = useContext(AuthContext);

  useEffect(() => {
    console.log('AuthState in AppRoutes:', authState);
  }, [authState]);

  if (authState.loading) {
    return <div>Loading...</div>;
  }

  return (
    <Routes>
      {/* Public Routes */}
      {!authState.authenticated && (
        <>
          <Route path="/" element={<SignIn />} />
          <Route path="/auth-redirect" element={<AuthRedirect />} />
          {/* Redirect any other routes to sign-in */}
          <Route path="*" element={<Navigate to="/" />} />
        </>
      )}

      {/* Protected Routes */}
      {authState.authenticated && (
        <>
          <Route path="/parentdashboard" element={<ParentDashboard />} />
          <Route path="/studentintake" element={<StudentIntake />} />
          <Route path="/booking" element={<BookingPage />} />
          {/* Redirect any other routes to dashboard */}
          <Route path="*" element={<Navigate to="/parentdashboard" />} />
        </>
      )}
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
