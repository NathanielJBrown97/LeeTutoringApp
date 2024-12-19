// src/components/StudentIntake.js

import React, { useState, useEffect } from 'react';
import { API_BASE_URL } from '../config';
import { useNavigate } from 'react-router-dom';
import './StudentIntake.css';

const StudentIntake = () => {
  const [numStudents, setNumStudents] = useState(1);
  const [studentIDs, setStudentIDs] = useState(['']);
  const [studentInfos, setStudentInfos] = useState([]);
  const [confirmationStep, setConfirmationStep] = useState(false);
  const [loading, setLoading] = useState(false);
  const [readyToProceed, setReadyToProceed] = useState(false);
  const navigate = useNavigate();

  const handleNumStudentsChange = (e) => {
    const count = parseInt(e.target.value) || 1;
    setNumStudents(count);
    setStudentIDs(Array(count).fill(''));
  };

  const handleStudentIDChange = (index, value) => {
    const newIDs = [...studentIDs];
    newIDs[index] = value;
    setStudentIDs(newIDs);
  };

  const handleSubmitStudentIDs = () => {
    setLoading(true);
    const token = localStorage.getItem('authToken');

    fetch(`${API_BASE_URL}/api/submitStudentIDs`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({ studentIds: studentIDs }),
    })
      .then(async (response) => {
        if (!response.ok) {
          const errorText = await response.text();
          throw new Error(`Error ${response.status}: ${errorText}`);
        }
        return response.json();
      })
      .then((data) => {
        console.log('Student Infos:', data);
        // Initialize 'confirmed' property to null
        const updatedInfos = data.studentInfos.map((info) => ({
          ...info,
          confirmed: null,
        }));
        setStudentInfos(updatedInfos);
        setConfirmationStep(true);
        setLoading(false);
      })
      .catch((error) => {
        console.error('Error submitting student IDs:', error);
        setLoading(false);
      });
  };

  const handleConfirmation = (index, confirmedValue) => {
    const updatedInfos = [...studentInfos];
    updatedInfos[index].confirmed = confirmedValue;
    setStudentInfos(updatedInfos);
  };

  useEffect(() => {
    // Check if all students have been confirmed/rejected
    if (confirmationStep) {
      const allConfirmed = studentInfos.every((info) => info.confirmed !== null);
      setReadyToProceed(allConfirmed);
    }
  }, [studentInfos, confirmationStep]);

  const handleProceed = () => {
    setLoading(true);
    const token = localStorage.getItem('authToken');

    // Extract confirmed student IDs
    const confirmedStudentIds = studentInfos
      .filter((info) => info.canLink && info.confirmed === true)
      .map((info) => info.studentId);

    // Proceed to link confirmed students
    fetch(`${API_BASE_URL}/api/confirmLinkStudents`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({ confirmedStudentIds }),
    })
      .then(async (response) => {
        if (!response.ok) {
          const errorText = await response.text();
          throw new Error(`Error ${response.status}: ${errorText}`);
        }
        return response.json();
      })
      .then((data) => {
        console.log('Students linked successfully:', data);
        setLoading(false);

        // Check if there are any students that were not confirmed
        const unconfirmedStudents = studentInfos.filter(
          (info) =>
            info.confirmed === false || (!info.canLink && info.confirmed !== true)
        );
        if (unconfirmedStudents.length > 0) {
          // Reset the intake form for unconfirmed students
          setNumStudents(unconfirmedStudents.length);
          setStudentIDs(Array(unconfirmedStudents.length).fill(''));
          setStudentInfos([]);
          setConfirmationStep(false);
          setReadyToProceed(false);
          alert('Some students were not confirmed. Please re-enter their IDs.');
        } else {
          // Redirect to Parent Dashboard
          navigate('/parentdashboard');
        }
      })
      .catch((error) => {
        console.error('Error confirming student links:', error);
        setLoading(false);
      });
  };

  return (
    <div className="student-intake-container">
      <h1>Associate Your Students</h1>
      {loading && <p>Loading...</p>}

      {!confirmationStep && (
        <div className="intake-form">
          <label>
            Number of Students:
            <input
              type="number"
              min="1"
              value={numStudents}
              onChange={handleNumStudentsChange}
            />
          </label>
          {studentIDs.map((id, index) => (
            <div key={index}>
              <label>
                Student ID {index + 1}:
                <input
                  type="text"
                  value={id}
                  onChange={(e) => handleStudentIDChange(index, e.target.value)}
                />
              </label>
            </div>
          ))}
          <button onClick={handleSubmitStudentIDs}>Submit Student IDs</button>
        </div>
      )}

      {confirmationStep && (
        <div className="confirmation-step">
          <h2>Confirm Your Students</h2>
          <ul>
            {studentInfos.map((info, index) => (
              <li key={index}>
                <p>
                  <strong>Student ID:</strong> {info.studentId}
                </p>
                <p>
                  <strong>Student Name:</strong> {info.studentName}
                </p>
                {!info.canLink && (
                  <p style={{ color: 'red' }}>
                    Cannot link this student. Please check the ID.
                  </p>
                )}
                {info.canLink && (
                  <div>
                    <p>Is this your student?</p>
                    <button
                      onClick={() => handleConfirmation(index, true)}
                      disabled={info.confirmed !== null}
                    >
                      Yes
                    </button>
                    <button
                      onClick={() => handleConfirmation(index, false)}
                      disabled={info.confirmed !== null}
                    >
                      No
                    </button>
                  </div>
                )}
                {info.confirmed === true && (
                  <p style={{ color: 'green' }}>Confirmed</p>
                )}
                {info.confirmed === false && (
                  <p style={{ color: 'orange' }}>Not your student</p>
                )}
              </li>
            ))}
          </ul>
          <button onClick={handleProceed} disabled={!readyToProceed}>
            Proceed
          </button>
        </div>
      )}
    </div>
  );
};

export default StudentIntake;