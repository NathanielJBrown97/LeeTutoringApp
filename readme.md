# Lee Tutoring LMS WebApp

## Disclaimer
This is an **active workspace**, meaning that features, functionality, and processes are subject to change as we continue to develop and enhance the platform. First and foremost; in order for the Portals below to eventually go live, the database (Google Firestore) must have an accurate import of all student data from the current systems. (Google Drive, Sheets, Calendars, Classrooms, and Copper). This is paramount, but has it's own challenges as everything was manually entered in the past and needs normalization to help ensure accuracy. 

## Sections Overview

### Parent Portal
The **Parent Portal** is designed to streamline parent involvement by offering a variety of key features:
- **Parent Intake**: Allowing new parents to easily sign up and link to their respective students.
- **Student Overview**: Displaying progress reports and relevant data on their students.
- **Handouts & Resources**: (Future feature) Providing additional learning resources and custom charts/graphics for better insight.
- **Hour Purchasing**: Handled through **Intuit QuickBooks** for smooth financial transactions.
- **Booking System**: Initially integrated with **YouCanBookMe** to manage appointments, but future updates will incorporate custom logic directly interfacing with **Google Calendar**, removing the need for third-party services.

### Tutor Portal
The **Tutor Portal** is not part of the initial minimum viable product but will be a significant component in later versions:
- **Student Management**: Tutors will have access to a roster of students, similar to the parent view but expanded for managing multiple students.
- **Lesson and Assignment Tools**: Integrating many of the tools currently used in Google Classroom, providing a robust environment for managing course material and tracking student progress.

### Student Portal
The **Student Portal** will be the final stage of development. This portal will serve as the primary hub for students to:
- **Student Login**: A dedicated route for students to access their courses, materials, and assignments.
- **Interaction Hub**: Students will be able to directly interact with Lee Tutoring for coursework, assignments, and potentially live sessions.
