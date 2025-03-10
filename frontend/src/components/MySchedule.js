import React, { useState, useEffect } from 'react';
import {
  Accordion,
  AccordionSummary,
  AccordionDetails,
  Typography,
  CircularProgress,
  Box,
  Card,
  CardContent
} from '@mui/material';
import { styled } from '@mui/system';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';

// -------------------- Styled Component for Appointment Card --------------------
const AppointmentCard = styled(Card)(({ theme }) => ({
  flex: '0 0 auto',
  boxShadow: '0 4px 12px rgba(0, 0, 0, 0.1)',
  borderRadius: theme.shape.borderRadius,
  borderLeft: `6px solid ${theme.palette.primary.main || '#b29600'}`,
  backgroundColor: '#fff',
  padding: theme.spacing(2),
  marginBottom: theme.spacing(2),
}));

// -------------------- Helper Functions --------------------

// Prevent auto-linking by inserting zero-width spaces.
function preventAutoLinking(str = '') {
  let safe = str.replace(/:\/\//g, ':\u200B//');
  safe = safe.replace(/\./g, '.\u200B');
  return safe;
}

/**
 * Extracts the "No Show / Reschedule / Cancel / Manage" links from raw HTML
 * so we can display them as separate buttons.
 */
function extractActionLinks(text) {
  const actions = {
    noShow: null,
    reschedule: null,
    cancel: null,
    manage: null
  };

  const noShowMatch = text.match(/<a\s+href=["']([^"']+)["'][^>]*>Mark as no show<\/a>/i);
  if (noShowMatch) actions.noShow = noShowMatch[1];

  const rescheduleMatch = text.match(/<a\s+href=["']([^"']+)["'][^>]*>Reschedule this booking<\/a>/i);
  if (rescheduleMatch) actions.reschedule = rescheduleMatch[1];

  const cancelMatch = text.match(/<a\s+href=["']([^"']+)["'][^>]*>Cancel this booking<\/a>/i);
  if (cancelMatch) actions.cancel = cancelMatch[1];

  const manageMatch = text.match(/<a\s+href=["']([^"']+)["'][^>]*>Manage this booking<\/a>/i);
  if (manageMatch) actions.manage = manageMatch[1];

  return actions;
}

/**
 * Forces certain field labels to start on a new line.
 */
function ensureFieldsOnNewLines(text) {
  const fieldLabels = [
    'Email:',
    'Parent email address (Optional):',
    'Phone number (Optional - if student is late):',
    'Type of Tutoring:',
    'Any Additional Info? (Optional):',
    'I understand that cancellation with less',
    'Appointment Type :',
    'Team member :',
    'YCBM link ref:'
  ];

  fieldLabels.forEach(label => {
    const re = new RegExp(`(?!^)(?=${label})`, 'g');
    text = text.replace(re, '\n');
  });
  return text;
}

/**
 * Extracts "Duration: ..." from the text.
 */
function extractDurationFromText(text) {
  const match = text.match(/(Duration:\s*.*)(\n|$)/i);
  let duration = null;
  if (match) {
    duration = match[1].replace(/Duration:\s*/i, '').trim();
    text = text.replace(match[0], '').trim();
  }
  return { duration, text };
}

/**
 * Cleans the raw HTML from the event description.
 */
function cleanDescription(text) {
  let cleaned = text
    .replace(/Mark as no show.*?(\n|<br\s*\/?>|$)/gi, '')
    .replace(/Reschedule this booking.*?(\n|<br\s*\/?>|$)/gi, '')
    .replace(/Cancel this booking.*?(\n|<br\s*\/?>|$)/gi, '')
    .replace(/Manage this booking.*?(\n|<br\s*\/?>|$)/gi, '')
    .replace(/Tutor Name:.*?(\n|<br\s*\/?>|$)/gi, '')
    .replace(/Tutor Email:.*?(\n|<br\s*\/?>|$)/gi, '')
    .replace(/Contact Your Tutor:.*?(\n|<br\s*\/?>|$)/gi, '')
    .replace(/First name:.*?(\n|<br\s*\/?>|$)/gi, '')
    .replace(/Last name:.*?(\n|<br\s*\/?>|$)/gi, '');

  cleaned = cleaned.replace(/<\/?p\b[^>]*>/gi, '\n');
  cleaned = cleaned.replace(/<br\s*\/?>/gi, '\n');
  cleaned = cleaned.replace(/<a\b[^>]*>([\s\S]*?)<\/a>/gi, '$1');
  cleaned = cleaned.replace(/<[^>]+>/g, '');
  cleaned = cleaned.trim();
  cleaned = preventAutoLinking(cleaned);
  cleaned = ensureFieldsOnNewLines(cleaned);

  return cleaned;
}

/**
 * Formats a given time object.
 */
function formatTime(timeObj) {
  if (timeObj?.dateTime) {
    return new Date(timeObj.dateTime).toLocaleTimeString([], {
      hour: '2-digit',
      minute: '2-digit',
    });
  } else if (timeObj?.date) {
    return new Date(timeObj.date).toLocaleTimeString([], {
      hour: '2-digit',
      minute: '2-digit',
    });
  }
  return '';
}

/**
 * A single event card component similar in style to Today's Schedule.
 */
function EventCard({ event }) {
  const [expanded, setExpanded] = useState(false);

  const toggleExpand = () => setExpanded(!expanded);

  const rawDescription = event.description || '';
  const actionLinks = extractActionLinks(rawDescription);
  let cleaned = cleanDescription(rawDescription);

  // Pull out "Duration" so we can show it separately.
  const { duration, text: remainder } = extractDurationFromText(cleaned);

  return (
    <AppointmentCard>
      <CardContent>
        <Typography variant="h6" sx={{ color: '#0e1027', fontWeight: 'bold', mb: 1 }}>
          {event.summary || 'No Title'}
        </Typography>

        <Typography variant="subtitle2" sx={{ color: '#b29600', mb: 1 }}>
          {formatTime(event.start)} - {formatTime(event.end)}
        </Typography>

        {event.location && (
          <Typography variant="body2" sx={{ mb: 1 }}>
            <a 
              href={event.location}
              target="_blank"
              rel="noopener noreferrer"
              style={{ color: '#0e1027', textDecoration: 'underline' }}
            >
              {event.location}
            </a>
          </Typography>
        )}

        {duration && (
          <Typography variant="body2" sx={{ color: '#0e1027', mb: 1, whiteSpace: 'pre-wrap' }}>
            <strong>Duration:</strong> {duration}
          </Typography>
        )}

        {remainder && (
          <>
            <Typography variant="body2" sx={{ color: '#0e1027', whiteSpace: 'pre-wrap' }}>
              {expanded ? remainder : ''}
            </Typography>
            <Typography
              onClick={toggleExpand}
              sx={{ mt: 1, textTransform: 'none', color: '#b29600', cursor: 'pointer' }}
            >
              {expanded ? 'Hide Details' : 'Show Details'}
            </Typography>
          </>
        )}
      </CardContent>
    </AppointmentCard>
  );
}

/**
 * MySchedule Component
 * Displays a weekly schedule as 7 Accordions (Sunday -> Saturday) for the current week.
 * All events for the week are fetched at once upon component mount for a smoother experience.
 * All accordions start collapsed.
 */
function MySchedule({ tutorId, backendUrl }) {
  const [weekDays, setWeekDays] = useState([]);
  const [dailySchedules, setDailySchedules] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [expanded, setExpanded] = useState(null); // No accordion expanded initially

  // Compute the current week (Sunday through Saturday) on mount.
  useEffect(() => {
    const today = new Date();
    const startOfWeek = new Date(today);
    startOfWeek.setDate(today.getDate() - today.getDay()); // Sunday as the start of the week
    const days = [];
    for (let i = 0; i < 7; i++) {
      const d = new Date(startOfWeek);
      d.setDate(startOfWeek.getDate() + i);
      days.push(d);
    }
    setWeekDays(days);
  }, []);

  // Once weekDays are set, fetch all calendar events once.
  useEffect(() => {
    async function fetchEvents() {
      setLoading(true);
      setError(null);
      try {
        const token = localStorage.getItem('authToken');
        const res = await fetch(
          `${backendUrl}/api/tutor/calendar-events?user_id=${encodeURIComponent(tutorId)}`,
          {
            method: 'GET',
            headers: {
              'Content-Type': 'application/json',
              'Authorization': 'Bearer ' + token,
            }
          }
        );
        if (!res.ok) {
          throw new Error(`Error: ${res.status}`);
        }
        const data = await res.json();
        const events = data.items || [];
        // For each day in the week, filter events that match the day.
        const schedules = weekDays.map(day => {
          return events.filter(event => {
            const start = event.start?.dateTime
              ? new Date(event.start.dateTime)
              : new Date(event.start.date);
            return (
              start.getFullYear() === day.getFullYear() &&
              start.getMonth() === day.getMonth() &&
              start.getDate() === day.getDate()
            );
          });
        });
        setDailySchedules(schedules);
      } catch (err) {
        setError(err.message);
      } finally {
        setLoading(false);
      }
    }
    if (tutorId && weekDays.length === 7) {
      fetchEvents();
    }
  }, [tutorId, backendUrl, weekDays]);

  const handleAccordionChange = (index) => (event, isExpanded) => {
    setExpanded(isExpanded ? index : null);
  };

  // Format header as "DayName M/D" (e.g., "Monday 3/10")
  const formatHeader = (date) => {
    const options = { weekday: 'long' };
    const dayName = date.toLocaleDateString('en-US', options);
    const month = date.getMonth() + 1;
    const day = date.getDate();
    return `${dayName} ${month}/${day}`;
  };

  return (
    <Box>
      {loading ? (
        <CircularProgress />
      ) : error ? (
        <Typography color="error">Error: {error}</Typography>
      ) : (
        weekDays.map((day, index) => (
          <Accordion key={index} expanded={expanded === index} onChange={handleAccordionChange(index)}>
            <AccordionSummary expandIcon={<ExpandMoreIcon />}>
              <Typography>{formatHeader(day)}</Typography>
            </AccordionSummary>
            <AccordionDetails>
              {dailySchedules[index] && dailySchedules[index].length > 0 ? (
                dailySchedules[index].map((event, idx) => (
                  <EventCard key={idx} event={event} />
                ))
              ) : (
                <Typography>No appointments for this day.</Typography>
              )}
            </AccordionDetails>
          </Accordion>
        ))
      )}
    </Box>
  );
}

export default MySchedule;
