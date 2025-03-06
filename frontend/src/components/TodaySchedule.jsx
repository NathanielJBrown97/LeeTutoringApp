import React, { useState, useEffect, useRef } from 'react';
import {
  Box,
  Typography,
  CircularProgress,
  Card,
  CardContent,
  Button,
  Collapse,
  IconButton
} from '@mui/material';
import { styled } from '@mui/system';
import ChevronLeftIcon from '@mui/icons-material/ChevronLeft';
import ChevronRightIcon from '@mui/icons-material/ChevronRight';

// -------------------- Brand Colors --------------------
const brandNavy = '#0e1027';
const brandGold = '#b29600';

// -------------------- Styled Components --------------------
const SliderContainer = styled(Box)(({ theme }) => ({
  position: 'relative',
  width: '100%',
  backgroundColor: brandNavy,
  overflowX: 'hidden',
  padding: theme.spacing(2),
}));

const ScrollContainer = styled(Box)(({ theme }) => ({
  display: 'flex',
  overflowX: 'auto',
  scrollBehavior: 'smooth',
  gap: theme.spacing(2),
  padding: theme.spacing(2, 6),
}));

const AppointmentCard = styled(Card)(({ theme }) => ({
  flex: '0 0 auto',
  boxShadow: '0 4px 12px rgba(0, 0, 0, 0.1)',
  borderRadius: theme.shape.borderRadius,
  borderLeft: `6px solid ${theme.palette.primary.main || brandGold}`,
  backgroundColor: '#fff',
  padding: theme.spacing(2),
}));

// -------------------- Helper Functions --------------------

// Prevent auto-linking by inserting zero-width spaces.
function preventAutoLinking(str = '') {
  let safe = str.replace(/:\/\//g, ':\u200B//');
  safe = safe.replace(/\./g, '.\u200B');
  return safe;
}

// Convert "Last, First" to "First Last"
function formatStudentName(name = '') {
  const parts = name.split(',');
  if (parts.length === 2) {
    return parts[1].trim() + ' ' + parts[0].trim();
  }
  return name;
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
 * A single Event card (one appointment).
 */
function EventCard({ event }) {
  const [expanded, setExpanded] = useState(false);

  const toggleExpand = () => setExpanded(!expanded);

  const formatTime = (timeObj) => {
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
  };

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

        <Typography variant="subtitle2" sx={{ color: brandGold, mb: 1 }}>
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

        <Box sx={{ display: 'flex', gap: 1, alignItems: 'center', mb: 1 }}>
          <Button
            variant="text"
            component="a"
            href={actionLinks.noShow || '#'}
            target="_blank"
            sx={{ textTransform: 'none', color: brandGold, p: 0, minWidth: 'auto' }}
          >
            No Show
          </Button>
          <Typography variant="body2" sx={{ color: brandGold }}>|</Typography>

          <Button
            variant="text"
            component="a"
            href={actionLinks.reschedule || '#'}
            target="_blank"
            sx={{ textTransform: 'none', color: brandGold, p: 0, minWidth: 'auto' }}
          >
            Reschedule
          </Button>
          <Typography variant="body2" sx={{ color: brandGold }}>|</Typography>

          <Button
            variant="text"
            component="a"
            href={actionLinks.cancel || '#'}
            target="_blank"
            sx={{ textTransform: 'none', color: brandGold, p: 0, minWidth: 'auto' }}
          >
            Cancel
          </Button>
          <Typography variant="body2" sx={{ color: brandGold }}>|</Typography>

          <Button
            variant="text"
            component="a"
            href={actionLinks.manage || '#'}
            target="_blank"
            sx={{ textTransform: 'none', color: brandGold, p: 0, minWidth: 'auto' }}
          >
            Manage
          </Button>
        </Box>

        {remainder && (
          <>
            <Collapse in={expanded}>
              <Box mt={1}>
                <Typography variant="body2" sx={{ color: '#0e1027', whiteSpace: 'pre-wrap' }}>
                  {remainder}
                </Typography>
              </Box>
            </Collapse>
            <Button
              onClick={toggleExpand}
              sx={{ mt: 1, textTransform: 'none', color: brandGold }}
            >
              {expanded ? 'Hide Details' : 'Show Details'}
            </Button>
          </>
        )}
      </CardContent>
    </AppointmentCard>
  );
}

/**
 * The main schedule component, used inside TutorDashboard.
 * It extracts all events for today (including past appointments) to build the student names array,
 * while still displaying only upcoming/ongoing events in the slider.
 */
function TodaySchedule({ tutorId, backendUrl, onStudentNamesUpdate }) {
  const [events, setEvents] = useState([]);
  const [studentNames, setStudentNames] = useState([]);
  const [loadingEvents, setLoadingEvents] = useState(true);
  const [error, setError] = useState(null);
  const scrollRef = useRef(null);

  useEffect(() => {
    async function fetchCalendarEvents() {
      try {
        const token = localStorage.getItem('authToken');
        const res = await fetch(
          `${backendUrl}/api/tutor/calendar-events?user_id=${encodeURIComponent(tutorId)}`,
          {
            method: 'GET',
            headers: {
              'Content-Type': 'application/json',
              'Authorization': 'Bearer ' + token
            }
          }
        );
        if (!res.ok) {
          throw new Error(`Fetch error, status: ${res.status}`);
        }
        const data = await res.json();
        const today = new Date();
        // Filter all events that occur today, regardless of whether they've ended
        const allTodayEvents = (data.items || []).filter(event => {
          const start = event.start?.dateTime
            ? new Date(event.start.dateTime)
            : new Date(event.start.date);
          return start.getFullYear() === today.getFullYear() &&
                 start.getMonth() === today.getMonth() &&
                 start.getDate() === today.getDate();
        });
        // For the slider, include only upcoming or ongoing events.
        const currentTime = new Date();
        const upcomingEvents = allTodayEvents.filter(event => {
          const start = event.start?.dateTime ? new Date(event.start.dateTime) : new Date(event.start.date);
          const end = event.end?.dateTime ? new Date(event.end.dateTime) : new Date(event.end.date);
          return start >= currentTime || (start <= currentTime && end > currentTime);
        });
        setEvents(upcomingEvents);
        // Extract student names from ALL today's events
        const names = allTodayEvents.map(event => {
          const summary = event.summary || 'No Name';
          return formatStudentName(summary);
        });
        setStudentNames(names);
        if (onStudentNamesUpdate) {
          onStudentNamesUpdate(names);
        }
      } catch (err) {
        console.error('Error fetching calendar events:', err);
        setError(err.message);
      } finally {
        setLoadingEvents(false);
      }
    }
    if (tutorId) {
      fetchCalendarEvents();
    }
  }, [tutorId, backendUrl, onStudentNamesUpdate]);

  const scrollLeft = () => {
    if (scrollRef.current) {
      scrollRef.current.scrollBy({ left: -window.innerWidth * 0.75, behavior: 'smooth' });
    }
  };

  const scrollRight = () => {
    if (scrollRef.current) {
      scrollRef.current.scrollBy({ left: window.innerWidth * 0.75, behavior: 'smooth' });
    }
  };

  if (loadingEvents) {
    return <CircularProgress />;
  }

  if (error) {
    return <Typography variant="body1" color="error">Error: {error}</Typography>;
  }

  return (
    <SliderContainer>
      {/* Left Chevron */}
      <IconButton
        onClick={scrollLeft}
        sx={{
          position: 'absolute',
          left: 0,
          top: '50%',
          transform: 'translateY(-50%)',
          zIndex: 1,
          color: brandGold,
          width: 64,
          height: 64
        }}
      >
        <ChevronLeftIcon sx={{ fontSize: 64 }} />
      </IconButton>

      {/* Scrollable Cards */}
      <ScrollContainer ref={scrollRef}>
        {events.length === 0 ? (
          <Typography variant="body1" sx={{ color: '#fff' }}>
            No upcoming or ongoing appointments for today.
          </Typography>
        ) : (
          events.map((event, idx) => (
            <EventCard key={idx} event={event} />
          ))
        )}
      </ScrollContainer>

      {/* Right Chevron */}
      <IconButton
        onClick={scrollRight}
        sx={{
          position: 'absolute',
          right: 0,
          top: '50%',
          transform: 'translateY(-50%)',
          zIndex: 1,
          color: brandGold,
          width: 64,
          height: 64
        }}
      >
        <ChevronRightIcon sx={{ fontSize: 64 }} />
      </IconButton>
    </SliderContainer>
  );
}

export default TodaySchedule;
