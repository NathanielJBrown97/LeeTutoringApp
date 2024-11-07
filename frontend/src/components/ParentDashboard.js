// src/components/ParentDashboard.js

import React, { useEffect, useState, useContext } from 'react';
import { API_BASE_URL } from '../config';
import { AuthContext } from '../contexts/AuthContext';
import './ParentDashboard.css'; // Optional: Create CSS for styling

const ParentDashboard = () => {
  const authState = useContext(AuthContext);
  const [dashboardData, setDashboardData] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetch(`${API_BASE_URL}/api/dashboard`, {
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
        setDashboardData(data);
        setLoading(false);
      })
      .catch((error) => {
        console.error('Error fetching dashboard data:', error);
        setLoading(false);
        // Optionally, set an error state here to display in the UI
      });
  }, []);

  const handleSignOut = () => {
    window.location.href = `${API_BASE_URL}/api/auth/signout`;
  };

  if (loading) {
    return <div>Loading Dashboard...</div>;
  }

  if (!dashboardData) {
    return <div>No Data Available</div>;
  }

  return (
    <div className="dashboard-container">
      <h1>Parent Dashboard</h1>
      <button onClick={handleSignOut} className="signout-button">Sign Out</button>
      <div className="welcome-section">
        <p>Welcome, <strong>{authState.user.email}</strong>!</p>
      </div>
      {/* You can add more dashboard-specific content here */}
    </div>
  );
};

export default ParentDashboard;
