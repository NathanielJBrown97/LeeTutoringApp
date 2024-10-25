// src/components/StudentIntake.js

import React, { useState } from 'react';
import axios from 'axios';
import { useNavigate } from 'react-router-dom';

const StudentIntake = () => {
  const [studentIds, setStudentIds] = useState(['']);
  const [studentInfos, setStudentInfos] = useState(null);
  const [error, setError] = useState('');
  const navigate = useNavigate();

  const handleAddStudentId = () => {
    setStudentIds([...studentIds, '']);
  };

  const handleStudentIdChange = (index, value) => {
    const newStudentIds = [...studentIds];
    newStudentIds[index] = value;
    setStudentIds(newStudentIds);
  };

  const handleSubmit = (e) => {
    e.preventDefault();

    // Remove empty IDs
    const idsToSubmit = studentIds.filter((id) => id.trim() !== '');

    if (idsToSubmit.length === 0) {
      setError('Please enter at least one student ID.');
      return;
    }

    // Send the student IDs to the backend
    axios
      .post('/api/submitStudentIDs', { studentIds: idsToSubmit }, { withCredentials: true })
      .then((response) => {
        setStudentInfos(response.data.studentInfos);
      })
      .catch((error) => {
        console.error('Error submitting student IDs:', error);
        setError('Error submitting student IDs.');
      });
  };

  const handleConfirm = () => {
    const confirmedStudentIds = studentInfos
      .filter((info) => info.canLink)
      .map((info) => info.studentId);

    // Send the confirmed student IDs to the backend
    axios
      .post('/api/confirmLinkStudents', { confirmedStudentIds }, { withCredentials: true })
      .then((response) => {
        // Redirect to the dashboard
        navigate('/parentdashboard');
      })
      .catch((error) => {
        console.error('Error confirming student IDs:', error);
        setError('Error confirming student IDs.');
      });
  };

  if (studentInfos) {
    return (
      <div style={styles.container}>
        <h1>Confirm Students</h1>
        {studentInfos.map((info, index) => (
          <div key={index}>
            {info.canLink ? (
              <p>
                Is {info.studentName} (ID: {info.studentId}) your child?
              </p>
            ) : (
              <p>
                {info.studentName} (ID: {info.studentId}) was not found or cannot be linked.
              </p>
            )}
          </div>
        ))}
        <button onClick={handleConfirm}>Confirm and Link Students</button>
      </div>
    );
  }

  return (
    <div style={styles.container}>
      <h1>Enter Student IDs</h1>
      {error && <p style={styles.error}>{error}</p>}
      <form onSubmit={handleSubmit}>
        {studentIds.map((id, index) => (
          <div key={index}>
            <input
              type="text"
              placeholder={`Student ID ${index + 1}`}
              value={id}
              onChange={(e) => handleStudentIdChange(index, e.target.value)}
            />
          </div>
        ))}
        <button type="button" onClick={handleAddStudentId}>
          Add Another Student ID
        </button>
        <br />
        <button type="submit">Submit Student IDs</button>
      </form>
    </div>
  );
};

const styles = {
  container: {
    margin: '2em',
  },
  error: {
    color: 'red',
  },
};

export default StudentIntake;
