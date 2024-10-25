// src/components/Dashboard.js

import React, { useEffect, useState } from 'react';
import axios from 'axios';
import { useNavigate } from 'react-router-dom';

const Dashboard = () => {
  const [dashboardData, setDashboardData] = useState(null);
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();

  useEffect(() => {
    // Fetch the dashboard data from the backend
    axios
      .get('/api/dashboard', { withCredentials: true })
      .then((response) => {
        if (response.data.needsStudentIntake) {
          // Redirect to the intake page
          navigate('/studentintake');
        } else {
          setDashboardData(response.data);
          setLoading(false);
        }
      })
      .catch((error) => {
        console.error('Error fetching dashboard data:', error);
        // Handle unauthorized access by redirecting to login
        if (error.response && error.response.status === 401) {
          window.location.href = '/';
        } else {
          setLoading(false);
        }
      });
  }, [navigate]);

  if (loading) {
    return <p>Loading dashboard...</p>;
  }

  if (!dashboardData) {
    return <p>Error loading dashboard data.</p>;
  }

  const {
    studentName,
    remainingHours,
    teamLead,
    associatedTutors = [],
    associatedStudents = [],
    recentActScores = [],
  } = dashboardData;

  return (
    <div style={styles.container}>
      <h1>Welcome to the Parent Dashboard</h1>
      <p>
        <strong>Student Name:</strong> {studentName}
      </p>
      <p>
        <strong>Remaining Hours:</strong> {remainingHours}
      </p>
      <p>
        <strong>Team Lead:</strong> {teamLead}
      </p>

      {associatedTutors.length > 0 && (
        <>
          <h2>Associated Tutors</h2>
          <ul>
            {associatedTutors.map((tutor, index) => (
              <li key={index}>{tutor}</li>
            ))}
          </ul>
        </>
      )}

      {associatedStudents.length > 0 && (
        <>
          <h2>Associated Students</h2>
          <ul>
            {associatedStudents.map((student) => (
              <li key={student.id}>{student.name}</li>
            ))}
          </ul>
        </>
      )}

      {recentActScores.length > 0 && (
        <>
          <h2>Recent ACT Scores</h2>
          <ul>
            {recentActScores.map((score, index) => (
              <li key={index}>{score}</li>
            ))}
          </ul>
        </>
      )}
    </div>
  );
};

const styles = {
  container: {
    margin: '2em',
  },
};

export default Dashboard;
