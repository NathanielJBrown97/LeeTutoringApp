// src/components/BookingPage.js

import React, { useEffect, useState, useContext } from 'react';
import { API_BASE_URL } from '../config';
import { useNavigate } from 'react-router-dom';
import { AuthContext } from '../contexts/AuthContext';
import './BookingPage.css';
import { tutorBookingLinks } from '../config/TutorBookingLinks'; // Import the tutor links mapping

const BookingPage = () => {
  const [studentsData, setStudentsData] = useState([]);
  const [loading, setLoading] = useState(true);
  const authState = useContext(AuthContext);
  const navigate = useNavigate();

  useEffect(() => {
    const token = localStorage.getItem('authToken');

    // Fetch associated students
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
        const associatedStudents = data.associatedStudents || [];
        if (associatedStudents.length > 0) {
          // Fetch data for each student
          fetchStudentsData(associatedStudents, token);
        } else {
          // If no associated students, redirect to intake
          navigate('/studentintake');
        }
      })
      .catch((error) => {
        console.error('Error fetching associated students:', error);
        setLoading(false);
      });
  }, [navigate]);

  const fetchStudentsData = (studentIDs, token) => {
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
        .then((studentData) => ({
          studentID,
          personal: studentData.personal || {},
          business: studentData.business || {},
        }))
    );

    Promise.all(fetchPromises)
      .then((students) => {
        setStudentsData(students);
        setLoading(false);
      })
      .catch((error) => {
        console.error('Error fetching student data:', error);
        setLoading(false);
      });
  };

  const handleBackToDashboard = () => {
    navigate('/parentdashboard');
  };

  if (loading) {
    return <div>Loading Booking Information...</div>;
  }

  return (
    <div className="booking-page-container">
      <h1>Book a Meeting</h1>
      <button onClick={handleBackToDashboard} className="back-button">
        Back to Dashboard
      </button>
      <div className="students-list">
        {studentsData.map((student) => (
          <div key={student.studentID} className="student-card">
            <h2>{student.personal.name || 'Unnamed Student'}</h2>
            <p>
              <strong>Team Lead:</strong> {student.business.team_lead || 'N/A'}
            </p>
            <p>
              <strong>Lifetime Hours:</strong>{' '}
              {student.business.lifetime_hours !== undefined
                ? parseFloat(student.business.lifetime_hours).toFixed(2)
                : 'N/A'}
            </p>
            <p>
              <strong>Remaining Hours:</strong> {student.business.remaining_hours || 'N/A'}
            </p>
            <p>
              <strong>Associated Tutors:</strong>
            </p>
            {student.business.associated_tutors &&
            Array.isArray(student.business.associated_tutors) ? (
              <div>
                {student.business.associated_tutors.map((tutor, index) => (
                  <div key={index} className="tutor-section">
                    <p>{tutor}</p>
                    {/* Check if tutor exists in the mapping */}
                    {tutorBookingLinks[tutor] ? (
                      <div className="booking-buttons">
                        <a
                          href={tutorBookingLinks[tutor]['1hr']}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="booking-button"
                        >
                          Book 1-Hour Session
                        </a>
                        <a
                          href={tutorBookingLinks[tutor]['30m']}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="booking-button"
                        >
                          Book 30-Minute Session
                        </a>
                      </div>
                    ) : (
                      <p>No booking links available for this tutor.</p>
                    )}
                  </div>
                ))}
              </div>
            ) : (
              <p>N/A</p>
            )}
          </div>
        ))}
      </div>
    </div>
  );
};

export default BookingPage;
