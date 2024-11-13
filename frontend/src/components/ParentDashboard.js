// src/components/ParentDashboard.js

import React, { useEffect, useState, useContext } from 'react';
import { API_BASE_URL } from '../config';
import { AuthContext } from '../contexts/AuthContext';
import { useNavigate } from 'react-router-dom';
import './ParentDashboard.css';

const ParentDashboard = () => {
  const authState = useContext(AuthContext);
  const [associatedStudents, setAssociatedStudents] = useState([]); // Array of student IDs
  const [studentsInfo, setStudentsInfo] = useState([]); // Array of student objects with id and name
  const [selectedStudentID, setSelectedStudentID] = useState(null);
  const [studentData, setStudentData] = useState(null);
  const [activeTab, setActiveTab] = useState('overview');
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();

  // Fetch associated students when the component mounts
  useEffect(() => {
    const token = localStorage.getItem('authToken');
    fetchAssociatedStudents(token);
  }, [navigate]);

  const fetchAssociatedStudents = (token) => {
    fetch(`${API_BASE_URL}/api/associated-students`, {
      method: 'GET',
      headers: {
        Authorization: `Bearer ${token}`,
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
        if (data.associatedStudents && data.associatedStudents.length > 0) {
          setAssociatedStudents(data.associatedStudents);
          // Fetch names of associated students
          fetchStudentsInfo(data.associatedStudents);
        } else {
          attemptAutomaticAssociation(token);
        }
      })
      .catch((error) => {
        console.error('Error fetching associated students:', error);
        setLoading(false);
      });
  };

  const attemptAutomaticAssociation = (token) => {
    fetch(`${API_BASE_URL}/api/attemptAutomaticAssociation`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
      .then(async (response) => {
        if (!response.ok) {
          navigate('/studentintake');
          return;
        }
        return response.json();
      })
      .then((data) => {
        if (data.associatedStudents && data.associatedStudents.length > 0) {
          setAssociatedStudents(data.associatedStudents);
          // Fetch names of associated students
          fetchStudentsInfo(data.associatedStudents);
        } else {
          navigate('/studentintake');
        }
      })
      .catch((error) => {
        console.error('Error during automatic association:', error);
        navigate('/studentintake');
      });
  };

  const fetchStudentsInfo = (studentIDs) => {
    const token = localStorage.getItem('authToken');
    const fetchPromises = studentIDs.map((studentID) =>
      fetch(`${API_BASE_URL}/api/students/${studentID}`, {
        method: 'GET',
        headers: {
          Authorization: `Bearer ${token}`,
        },
      })
        .then(async (response) => {
          if (!response.ok) {
            const errorText = await response.text();
            throw new Error(`Error ${response.status}: ${errorText}`);
          }
          return response.json();
        })
        .then((data) => ({
          id: studentID,
          name: data.personal.name || studentID, // Use ID if name is not available
        }))
    );

    Promise.all(fetchPromises)
      .then((students) => {
        setStudentsInfo(students);
        setSelectedStudentID(students[0].id);
        setLoading(false);
      })
      .catch((error) => {
        console.error('Error fetching students info:', error);
        setLoading(false);
      });
  };

  // Fetch detailed student data when a student is selected
  useEffect(() => {
    if (selectedStudentID) {
      const token = localStorage.getItem('authToken');
      setLoading(true);
      fetch(`${API_BASE_URL}/api/students/${selectedStudentID}`, {
        method: 'GET',
        headers: {
          Authorization: `Bearer ${token}`,
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
    localStorage.removeItem('authToken');
    authState.updateToken(null);
    window.location.href = '/';
  };

  const handleStudentChange = (event) => {
    setSelectedStudentID(event.target.value);
  };

  const handleBookMeeting = () => {
    navigate('/booking');
  };

  const handleTabChange = (tabName) => {
    setActiveTab(tabName);
  };

  if (loading) {
    return <div className="loading">Loading Dashboard...</div>;
  }

  if (!studentData) {
    return (
      <div className="dashboard-container">
        <nav className="navbar">
          <h1 className="logo">Agora Parent Dashboard</h1>
          <div className="nav-buttons">
            <button onClick={handleBookMeeting} className="nav-button">
              Book a Meeting
            </button>
            <button onClick={handleSignOut} className="nav-button signout">
              Sign Out
            </button>
          </div>
        </nav>
        <div className="content">
          <div className="welcome-section">
            <p>
              Welcome, <strong>{authState.user.email}</strong>!
            </p>
            <p>Select a student to view their details.</p>
            <div className="student-selector">
              <label htmlFor="student-select">Select Student:</label>
              <select
                id="student-select"
                onChange={handleStudentChange}
                value={selectedStudentID}
              >
                {studentsInfo.map((student) => (
                  <option key={student.id} value={student.id}>
                    {student.name}
                  </option>
                ))}
              </select>
            </div>
          </div>
        </div>
      </div>
    );
  }

  // Extract student information
  const personalInfo = studentData.personal || {};
  const businessInfo = studentData.business || {};
  const goals = studentData.goals || [];
  const testDates = studentData.testDates || [];
  const homeworkCompletion = studentData.homeworkCompletion || [];
  const testData = studentData.testData || [];

  // Format lifetime_hours
  const formattedLifetimeHours =
    businessInfo.lifetime_hours !== undefined
      ? parseFloat(businessInfo.lifetime_hours).toFixed(2)
      : 'N/A';

  // Get selected student's name
  const selectedStudent = studentsInfo.find((s) => s.id === selectedStudentID);
  const selectedStudentName = selectedStudent ? selectedStudent.name : 'Student';

  return (
    <div className="dashboard-container">
      <nav className="navbar">
        <h1 className="logo">Agora Parent Dashboard</h1>
        <div className="nav-buttons">
          <button onClick={handleBookMeeting} className="nav-button">
            Book a Meeting
          </button>
          <button onClick={handleSignOut} className="nav-button signout">
            Sign Out
          </button>
        </div>
      </nav>
      <div className="content">
        <div className="student-selector">
          <label htmlFor="student-select">Select Student:</label>
          <select
            id="student-select"
            onChange={handleStudentChange}
            value={selectedStudentID}
          >
            {studentsInfo.map((student) => (
              <option key={student.id} value={student.id}>
                {student.name}
              </option>
            ))}
          </select>
        </div>

        <h2 className="student-name">Viewing: {selectedStudentName}</h2>

        <div className="tabs">
          <button
            className={`tab-button ${activeTab === 'overview' ? 'active' : ''}`}
            onClick={() => handleTabChange('overview')}
          >
            Overview
          </button>
          <button
            className={`tab-button ${activeTab === 'homework' ? 'active' : ''}`}
            onClick={() => handleTabChange('homework')}
          >
            Homework
          </button>
          <button
            className={`tab-button ${activeTab === 'tests' ? 'active' : ''}`}
            onClick={() => handleTabChange('tests')}
          >
            Tests
          </button>
        </div>

        <div className="tab-content">
          {activeTab === 'overview' && (
            <div className="overview-tab">
              <h2>Student Overview</h2>
              <table className="info-table">
                <tbody>
                  <tr>
                    <th colSpan="2">Personal Information</th>
                  </tr>
                  <tr>
                    <td>Name</td>
                    <td>{personalInfo.name || 'N/A'}</td>
                  </tr>
                  <tr>
                    <td>Accommodations</td>
                    <td>{personalInfo.accommodations || 'N/A'}</td>
                  </tr>
                  <tr>
                    <td>Grade</td>
                    <td>{personalInfo.grade || 'N/A'}</td>
                  </tr>
                  <tr>
                    <td>High School</td>
                    <td>{personalInfo.high_school || 'N/A'}</td>
                  </tr>
                  <tr>
                    <th colSpan="2">Business Information</th>
                  </tr>
                  <tr>
                    <td>Lifetime Hours</td>
                    <td>{formattedLifetimeHours}</td>
                  </tr>
                  <tr>
                    <td>Remaining Hours</td>
                    <td>{businessInfo.remaining_hours || 'N/A'}</td>
                  </tr>
                  <tr>
                    <td>Registered Tests</td>
                    <td>{businessInfo.registered_tests || 'N/A'}</td>
                  </tr>
                  <tr>
                    <td>Status</td>
                    <td>{businessInfo.status || 'N/A'}</td>
                  </tr>
                  <tr>
                    <td>Team Lead</td>
                    <td>{businessInfo.team_lead || 'N/A'}</td>
                  </tr>
                  <tr>
                    <td>Test Focus</td>
                    <td>{businessInfo.test_focus || 'N/A'}</td>
                  </tr>
                  <tr>
                    <td>Associated Tutors</td>
                    <td>
                      {businessInfo.associated_tutors &&
                      Array.isArray(businessInfo.associated_tutors) ? (
                        <ul>
                          {businessInfo.associated_tutors.map((tutor, index) => (
                            <li key={index}>{tutor}</li>
                          ))}
                        </ul>
                      ) : (
                        'N/A'
                      )}
                    </td>
                  </tr>
                  <tr>
                    <th colSpan="2">Goals</th>
                  </tr>
                  {goals.length > 0 ? (
                    goals.map((goal, index) => (
                      <React.Fragment key={index}>
                        <tr>
                          <td>College</td>
                          <td>{goal.College || 'N/A'}</td>
                        </tr>
                        {goal.ACT_percentiles && (
                          <>
                            <tr>
                              <td>ACT 25th Percentile</td>
                              <td>{goal.ACT_percentiles[0]}</td>
                            </tr>
                            <tr>
                              <td>ACT 50th Percentile</td>
                              <td>{goal.ACT_percentiles[1]}</td>
                            </tr>
                            <tr>
                              <td>ACT 75th Percentile</td>
                              <td>{goal.ACT_percentiles[2]}</td>
                            </tr>
                          </>
                        )}
                        {goal.SAT_percentiles && (
                          <>
                            <tr>
                              <td>SAT 25th Percentile</td>
                              <td>{goal.SAT_percentiles[0]}</td>
                            </tr>
                            <tr>
                              <td>SAT 50th Percentile</td>
                              <td>{goal.SAT_percentiles[1]}</td>
                            </tr>
                            <tr>
                              <td>SAT 75th Percentile</td>
                              <td>{goal.SAT_percentiles[2]}</td>
                            </tr>
                          </>
                        )}
                        <tr className="goal-divider">
                          <td colSpan="2"></td>
                        </tr>
                      </React.Fragment>
                    ))
                  ) : (
                    <tr>
                      <td colSpan="2">No goals available.</td>
                    </tr>
                  )}
                  <tr>
                    <th colSpan="2">Test Dates</th>
                  </tr>
                  {testDates.length > 0 ? (
                    testDates.map((date, index) => (
                      <React.Fragment key={index}>
                        <tr>
                          <td>Test Type</td>
                          <td>{date['Test Type'] || 'N/A'}</td>
                        </tr>
                        <tr>
                          <td>Test Date</td>
                          <td>{date['Test Date'] || 'N/A'}</td>
                        </tr>
                        <tr>
                          <td>Registration Deadline</td>
                          <td>{date['Regular Registration Deadline'] || 'N/A'}</td>
                        </tr>
                        <tr>
                          <td>Late Registration Deadline</td>
                          <td>{date['Late Registration Deadline'] || 'N/A'}</td>
                        </tr>
                        <tr>
                          <td>Notes</td>
                          <td>{date['Notes'] || 'N/A'}</td>
                        </tr>
                        <tr className="test-date-divider">
                          <td colSpan="2"></td>
                        </tr>
                      </React.Fragment>
                    ))
                  ) : (
                    <tr>
                      <td colSpan="2">No test dates available.</td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          )}

          {activeTab === 'homework' && (
            <div className="homework-tab">
              <h2>Homework Completion</h2>
              {homeworkCompletion.length > 0 ? (
                <table className="info-table">
                  <thead>
                    <tr>
                      <th>Date</th>
                      <th>Percentage Completed</th>
                      <th>Status</th>
                    </tr>
                  </thead>
                  <tbody>
                    {homeworkCompletion.map((hw, index) => {
                      const date = hw.date || hw.id || 'N/A';
                      const percentage =
                        hw.percentage !== undefined ? hw.percentage : 'N/A';
                      let status = 'N/A';

                      if (percentage === 0) {
                        status = 'Not Completed';
                      } else if (percentage === 100) {
                        status = 'Completed';
                      } else if (percentage > 0 && percentage < 100) {
                        status = 'Partially Completed';
                      }

                      return (
                        <tr key={index}>
                          <td>{date}</td>
                          <td>{percentage}%</td>
                          <td>{status}</td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              ) : (
                <p>No homework records available.</p>
              )}
            </div>
          )}

          {activeTab === 'tests' && (
            <div className="tests-tab">
              <h2>Test Data</h2>
              {testData.length > 0 ? (
                testData.map((test, index) => (
                  <div key={index} className="test-card">
                    <h3>{test.id || 'N/A'}</h3>
                    <p>
                      <strong>Date:</strong> {test.Date || 'N/A'}
                    </p>
                    <p>
                      <strong>Test:</strong> {test.Test || 'N/A'}
                    </p>
                    <p>
                      <strong>Baseline:</strong>{' '}
                      {test.Baseline !== undefined ? test.Baseline.toString() : 'N/A'}
                    </p>
                    <p>
                      <strong>Type:</strong> {test.Type || 'N/A'}
                    </p>
                    {/* ACT Scores */}
                    {test.ACT && (
                      <div>
                        <h4>ACT Scores</h4>
                        {test.ACT['ACT Total'] !== undefined && (
                          <p>
                            <strong>Total:</strong> {test.ACT['ACT Total']}
                          </p>
                        )}
                        {['English', 'Math', 'Reading', 'Science'].map((section) => (
                          <p key={section}>
                            <strong>{section}:</strong>{' '}
                            {test.ACT[section] !== undefined ? test.ACT[section] : 'N/A'}
                          </p>
                        ))}
                      </div>
                    )}
                    {/* SAT Scores */}
                    {test.SAT && (
                      <div>
                        <h4>SAT Scores</h4>
                        {test.SAT['SAT Total'] !== undefined && (
                          <p>
                            <strong>Total:</strong> {test.SAT['SAT Total']}
                          </p>
                        )}
                        {['EBRW', 'Math'].map((section) => (
                          <p key={section}>
                            <strong>
                              {section === 'EBRW'
                                ? 'Evidence-Based Reading and Writing'
                                : section}
                              :
                            </strong>{' '}
                            {test.SAT[section] !== undefined ? test.SAT[section] : 'N/A'}
                          </p>
                        ))}
                      </div>
                    )}
                  </div>
                ))
              ) : (
                <p>No test data available.</p>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default ParentDashboard;
