// src/components/ParentProfile.js

import React, { useEffect, useState } from 'react';
import { API_BASE_URL } from '../config';
import './ParentProfile.css'; // Optional: Create a CSS file for styling

const ParentProfile = () => {
  const [parentData, setParentData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    const fetchParentData = async () => {
      const token = localStorage.getItem('authToken');

      if (!token) {
        setError('No authentication token found. Please sign in.');
        setLoading(false);
        return;
      }

      try {
        const response = await fetch(`${API_BASE_URL}/api/parent`, {
          method: 'GET',
          headers: {
            Authorization: `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
        });

        if (!response.ok) {
          const errorText = await response.text();
          throw new Error(`Error ${response.status}: ${errorText}`);
        }

        const data = await response.json();
        setParentData(data);
      } catch (err) {
        console.error('Error fetching parent data:', err);
        setError(err.message);
      } finally {
        setLoading(false);
      }
    };

    fetchParentData();
  }, []);

  if (loading) {
    return <div>Loading parent profile...</div>;
  }

  if (error) {
    return <div style={{ color: 'red' }}>Error: {error}</div>;
  }

  if (!parentData) {
    return <div>No parent data available.</div>;
  }

  return (
    <div className="parent-profile-container">
      <h2>Parent Profile</h2>
      <div className="profile-info">
        {parentData.picture ? (
          <img
            src={parentData.picture}
            alt={`${parentData.name}'s profile`}
            className="profile-picture"
          />
        ) : (
          <div className="placeholder-picture">No Image</div>
        )}
        <div className="profile-details">
          <p>
            <strong>Name:</strong> {parentData.name || 'N/A'}
          </p>
          <p>
            <strong>Email:</strong> {parentData.email || 'N/A'}
          </p>
          <p>
            <strong>User ID:</strong> {parentData.user_id || 'N/A'}
          </p>
        </div>
      </div>
    </div>
  );
};

export default ParentProfile;
