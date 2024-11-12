// src/components/ParentDashboard.js

import React, { useEffect, useState, useContext } from 'react';
import { API_BASE_URL } from '../config';
import { AuthContext } from '../contexts/AuthContext';
import './ParentDashboard.css';

const ParentDashboard = () => {
  const authState = useContext(AuthContext);
  const [associatedStudents, setAssociatedStudents] = useState([]);
  const [selectedStudentID, setSelectedStudentID] = useState(null);
  const [studentData, setStudentData] = useState(null);
  const [loading, setLoading] = useState(true);

  // Fetch the associated students when the component mounts
  useEffect(() => {
    const token = localStorage.getItem('authToken');
    console.log('Token for API request:', token);

    // Fetch associated students
    fetch(`${API_BASE_URL}/api/associated-students`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${token}`,
      },
    })
      .then(async (response) => {
        if (!response.ok) {
          const errorText = await response.text();
          throw new Error(`Error ${response.status}: ${errorText}`);
        }
        return response.json();
      })
      .then((data) => {
        console.log('Associated students:', data);
        setAssociatedStudents(data.associatedStudents || []);
        if (data.associatedStudents && data.associatedStudents.length > 0) {
          // Set the first student as the selected student
          setSelectedStudentID(data.associatedStudents[0]);
        } else {
          setLoading(false);
        }
      })
      .catch((error) => {
        console.error('Error fetching associated students:', error);
        setLoading(false);
      });
  }, []);

  // Fetch detailed student data when a student is selected
  useEffect(() => {
    if (selectedStudentID) {
      const token = localStorage.getItem('authToken');
      setLoading(true);
      fetch(`${API_BASE_URL}/api/students/${selectedStudentID}`, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      })
        .then(async (response) => {
          if (!response.ok) {
            const errorText = await response.text();
            throw new Error(`Error ${response.status}: ${errorText}`);
          }
          return response.json();
        })
        .then((data) => {
          console.log('Detailed student data:', data);
          setStudentData(data);
          setLoading(false);
        })
        .catch((error) => {
          console.error('Error fetching student details:', error);
          setLoading(false);
        });
    }
  }, [selectedStudentID]);

  const handleSignOut = () => {
    // Remove the token from localStorage
    localStorage.removeItem('authToken');
    // Update the authentication context
    authState.updateToken(null);
    // Redirect to sign-in page
    window.location.href = '/';
  };

  const handleStudentChange = (event) => {
    setSelectedStudentID(event.target.value);
  };

  if (loading) {
    return <div>Loading Dashboard...</div>;
  }

  if (!associatedStudents.length) {
    return (
      <div className="dashboard-container">
        <h1>Parent Dashboard</h1>
        <button onClick={handleSignOut} className="signout-button">
          Sign Out
        </button>
        <div className="welcome-section">
          <p>
            Welcome, <strong>{authState.user.email}</strong>!
          </p>
          <p>No associated students found.</p>
        </div>
      </div>
    );
  }

  if (!studentData) {
    return (
      <div className="dashboard-container">
        <h1>Parent Dashboard</h1>
        <button onClick={handleSignOut} className="signout-button">
          Sign Out
        </button>
        <div className="welcome-section">
          <p>
            Welcome, <strong>{authState.user.email}</strong>!
          </p>
          <p>Select a student to view their details.</p>
          <select onChange={handleStudentChange} value={selectedStudentID}>
            {associatedStudents.map((studentID) => (
              <option key={studentID} value={studentID}>
                {studentID}
              </option>
            ))}
          </select>
        </div>
      </div>
    );
  }

  // Access the personal and business information
  const personalInfo = studentData.personal || {};
  const businessInfo = studentData.business || {};

  // Format lifetime_hours to two decimal places
  const formattedLifetimeHours = businessInfo.lifetime_hours !== undefined
    ? parseFloat(businessInfo.lifetime_hours).toFixed(2)
    : 'N/A';

  return (
    <div className="dashboard-container">
      <h1>Parent Dashboard</h1>
      <button onClick={handleSignOut} className="signout-button">
        Sign Out
      </button>
      <div className="welcome-section">
        <p>
          Welcome, <strong>{authState.user.email}</strong>!
        </p>
        <div className="student-selector">
          <label htmlFor="student-select">Select Student:</label>
          <select
            id="student-select"
            onChange={handleStudentChange}
            value={selectedStudentID}
          >
            {associatedStudents.map((studentID) => (
              <option key={studentID} value={studentID}>
                {studentID}
              </option>
            ))}
          </select>
        </div>
      </div>

      <div className="student-details">
        <h2>Personal Information</h2>
        <p>
          <strong>Name:</strong> {personalInfo.name || 'N/A'}
        </p>
        <p>
          <strong>Accommodations:</strong> {personalInfo.accommodations || 'N/A'}
        </p>
        <p>
          <strong>Grade:</strong> {personalInfo.grade || 'N/A'}
        </p>
        <p>
          <strong>High School:</strong> {personalInfo.high_school || 'N/A'}
        </p>
        <p>
          <strong>Parent Email:</strong> {personalInfo.parent_email || 'N/A'}
        </p>
        <p>
          <strong>Student Email:</strong> {personalInfo.student_email || 'N/A'}
        </p>

        <h2>Business Information</h2>
        <p>
          <strong>Lifetime Hours:</strong> {formattedLifetimeHours}
        </p>
        <p>
          <strong>Remaining Hours:</strong> {businessInfo.remaining_hours || 'N/A'}
        </p>
        <p>
          <strong>Registered Tests:</strong> {businessInfo.registered_tests || 'N/A'}
        </p>
        <p>
          <strong>Status:</strong> {businessInfo.status || 'N/A'}
        </p>
        <p>
          <strong>Team Lead:</strong> {businessInfo.team_lead || 'N/A'}
        </p>
        <p>
          <strong>Test Focus:</strong> {businessInfo.test_focus || 'N/A'}
        </p>
        <p>
          <strong>Associated Tutors:</strong>
        </p>
        {businessInfo.associated_tutors && Array.isArray(businessInfo.associated_tutors) ? (
          <ul>
            {businessInfo.associated_tutors.map((tutor, index) => (
              <li key={index}>{tutor}</li>
            ))}
          </ul>
        ) : (
          <p>N/A</p>
        )}

        <h2>Homework Completion</h2>
        {studentData.homeworkCompletion && studentData.homeworkCompletion.length > 0 ? (
          <ul>
            {studentData.homeworkCompletion.map((hw) => {
              const date = hw.date || hw.id || 'N/A';
              const percentage = hw.percentage !== undefined ? hw.percentage : 'N/A';
              let status = 'N/A';

              if (percentage === 0) {
                status = 'Not Completed';
              } else if (percentage === 100) {
                status = 'Completed';
              } else if (percentage > 0 && percentage < 100) {
                status = 'Partially Completed';
              }

              return (
                <li key={hw.id}>
                  <strong>Assignment Date:</strong> {date}<br />
                  <strong>Percentage Completed:</strong> {percentage}%<br />
                  <strong>Status:</strong> {status}
                </li>
              );
            })}
          </ul>
        ) : (
          <p>No homework records available.</p>
        )}

        <h2>Test Data</h2>
        {studentData.testData && studentData.testData.length > 0 ? (
          <div>
            {studentData.testData.map((test) => (
              <div key={test.id} className="test-data-entry">
                <h3>{test.id || 'N/A'}</h3>
                <p><strong>Date:</strong> {test.Date || 'N/A'}</p>
                <p><strong>Test:</strong> {test.Test || 'N/A'}</p>
                <p><strong>Baseline:</strong> {test.Baseline !== undefined ? test.Baseline.toString() : 'N/A'}</p>
                <p><strong>Type:</strong> {test.Type || 'N/A'}</p>

                {/* ACT Scores */}
                {test.ACT && (
                  <div>
                    <h4>ACT Scores</h4>
                    {test.ACT['ACT Total'] !== undefined && (
                      <p><strong>Total:</strong> {test.ACT['ACT Total']}</p>
                    )}
                    {/* Check if any section scores are present */}
                    {['English', 'Math', 'Reading', 'Science'].some(
                      (section) => test.ACT[section] !== undefined
                    ) && (
                      <>
                        {['English', 'Math', 'Reading', 'Science'].map((section) => {
                          if (test.ACT[section] !== undefined) {
                            return (
                              <p key={section}>
                                <strong>{section}:</strong> {test.ACT[section]}
                              </p>
                            );
                          } else {
                            return (
                              <p key={section}>
                                <strong>{section}:</strong> N/A
                              </p>
                            );
                          }
                        })}
                      </>
                    )}
                  </div>
                )}

                {/* SAT Scores */}
                {test.SAT && (
                  <div>
                    <h4>SAT Scores</h4>
                    {test.SAT['SAT Total'] !== undefined && (
                      <p><strong>Total:</strong> {test.SAT['SAT Total']}</p>
                    )}
                    {/* Check if any section scores are present */}
                    {['EBRW', 'Math'].some(
                      (section) => test.SAT[section] !== undefined
                    ) && (
                      <>
                        {['EBRW', 'Math'].map((section) => {
                          if (test.SAT[section] !== undefined) {
                            return (
                              <p key={section}>
                                <strong>
                                  {section === 'EBRW' ? 'Evidence-Based Reading and Writing' : section}:
                                </strong> {test.SAT[section]}
                              </p>
                            );
                          } else {
                            return (
                              <p key={section}>
                                <strong>
                                  {section === 'EBRW' ? 'Evidence-Based Reading and Writing' : section}:
                                </strong> N/A
                              </p>
                            );
                          }
                        })}
                      </>
                    )}
                  </div>
                )}
              </div>
            ))}
          </div>
        ) : (
          <p>No test data available.</p>
        )}

        <h2>Test Dates</h2>
        {studentData.testDates && studentData.testDates.length > 0 ? (
          <div>
            {studentData.testDates.map((date) => (
              <div key={date.id} className="test-date-entry">
                <h3>{date.id || 'N/A'}</h3>
                <p><strong>Test Type:</strong> {date['Test Type'] || 'N/A'}</p>
                <p><strong>Test Date:</strong> {date['Test Date'] || 'N/A'}</p>
                <p><strong>Registration Deadline:</strong> {date['Regular Registration Deadline'] || 'N/A'}</p>
                <p><strong>Late Registration Deadline:</strong> {date['Late Registration Deadline'] || 'N/A'}</p>
                <p><strong>Notes:</strong> {date['Notes'] || 'N/A'}</p>
              </div>
            ))}
          </div>
        ) : (
          <p>No test dates available.</p>
        )}

        <h2>Goals</h2>
        {studentData.goals && studentData.goals.length > 0 ? (
          <div>
            {studentData.goals.map((goal) => (
              <div key={goal.id} className="goal-entry">
                <h3>{goal.College || 'N/A'}</h3>
                {goal.ACT_percentiles && goal.ACT_percentiles.length >= 3 && (
                  <div>
                    <h4>ACT Percentiles</h4>
                    <p><strong>25th Percentile:</strong> {goal.ACT_percentiles[0]}</p>
                    <p><strong>50th Percentile:</strong> {goal.ACT_percentiles[1]}</p>
                    <p><strong>75th Percentile:</strong> {goal.ACT_percentiles[2]}</p>
                  </div>
                )}
                {goal.SAT_percentiles && goal.SAT_percentiles.length >= 3 && (
                  <div>
                    <h4>SAT Percentiles</h4>
                    <p><strong>25th Percentile:</strong> {goal.SAT_percentiles[0]}</p>
                    <p><strong>50th Percentile:</strong> {goal.SAT_percentiles[1]}</p>
                    <p><strong>75th Percentile:</strong> {goal.SAT_percentiles[2]}</p>
                  </div>
                )}
              </div>
            ))}
          </div>
        ) : (
          <p>No goals available.</p>
        )}
      </div>
    </div>
  );
};

export default ParentDashboard;
