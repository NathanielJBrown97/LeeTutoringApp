// src/App.js

import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import Login from './components/Login';
import Dashboard from './components/Dashboard';
import StudentIntake from './components/StudentIntake';

function App() {
  return (
    <Router>
      <Routes>
        <Route path="/" element={<Login />} />
        <Route path="/parentdashboard" element={<Dashboard />} />
        <Route path="/studentintake" element={<StudentIntake />} />
      </Routes>
    </Router>
  );
}

export default App;
