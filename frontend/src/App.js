// src/App.js

import React, { useContext, useEffect } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider, AuthContext } from './contexts/AuthContext';
import SignIn from './components/SignIn';
import ParentDashboard from './components/ParentDashboard';
import TutorDashboard from './components/TutorDashboard';
import StudentDashboard from './components/StudentDashboard';
import StudentIntake from './components/StudentIntake';
import BookingPage from './components/BookingPage';
import AuthRedirect from './components/AuthRedirect';
import ParentProfile from './components/ParentProfile';
import NoScrollWrapper from './components/NoScrollWrapper';

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
      {/* Public Routes (not logged in) */}
      {!authState.authenticated && (
        <>
          <Route
            path="/"
            element={
              <NoScrollWrapper>
                <SignIn />
              </NoScrollWrapper>
            }
          />
          <Route path="/auth-redirect" element={<AuthRedirect />} />
          <Route path="*" element={<Navigate to="/" />} />
        </>
      )}

      {/* Private Routes (logged in) */}
      {authState.authenticated && (
        <>
          <Route path="/parentdashboard" element={<ParentDashboard />} />
          <Route path="/tutordashboard" element={<TutorDashboard />} />
          <Route path="/studentdashboard" element={<StudentDashboard />} />
          <Route path="/studentintake" element={<StudentIntake />} />
          <Route path="/booking" element={<BookingPage />} />
          <Route path="/parentprofile" element={<ParentProfile />} />
          {/* Default route based on role */}
          <Route
            path="*"
            element={
              <Navigate
                to={
                  authState.user?.role === 'tutor'
                    ? '/tutordashboard'
                    : authState.user?.role === 'student'
                    ? '/studentdashboard'
                    : '/parentdashboard'
                }
              />
            }
          />
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
