// https://docs.google.com/forms/d/e/1FAIpQLSfZzCchWO7I0OjoKCXOWEEmiHPyd7PvKSGcTpg4GyUVFuoy7w/viewform?usp=pp_url&entry.1461756543=Lee&entry.1741532836=Edward&entry.1122527626=theelee13@gmail.com&entry.976932403=---&entry.89953784=---&entry.1176647490=12&entry.1902591832=TBD&entry.1097406917=Unknown&entry.1187838221=No

function onFormSubmit(e) {
    var responses = e.response.getItemResponses();
    var respondentEmail = responses[0].getResponse();
    Logger.log("respondent: " + respondentEmail);
    for(var i =0;i<responses.length;i++){
      Logger.log(responses[i].getResponse());
    }
    let student_data = {
                      "first_name": responses[2].getResponse(),
                      "last_name": responses[1].getResponse(),
                      "student_email": responses[3].getResponse(),
                      "student_number":responses[4].getResponse(),
                      "parent_email": responses[5].getResponse(),
                      "parent_number": responses[6].getResponse(),
                      "scheduler": responses[7].getResponse(),
                      "school": responses[8].getResponse(),
                      "grade": responses[9].getResponse(),
                      "test_focus": responses[10].getResponse(),
                      "accommodations": responses[11].getResponse(),
                      "interests": responses[12].getResponse(),
                      "availability": responses[13].getResponse(),
    }
    Logger.log(student_data.school);
    let registered = responses[14].getResponse();
    if(registered=="Yes"){
      Logger.log("Student has registered for a test");
      student_data["registered_tests"]=responses[15].getResponse();
    }
    else{
      Logger.log("No upcoming tests logged for student");
    }
      
    createNewStudent(student_data, respondentEmail);
  
    // After New Student is Created in Sheets ---> Send same data to our server to create student in Firestore DB.
    sendStudentDataToBackend(student_data);
  
  }
  
  function createNewStudent(e, respondentEmail) {
    var tutFolder = DriveApp.getFolderById("16A9u0AW4DLRcXfBQ4TrEsfQyYuJP3RRK");
    var studentFolder = tutFolder.createFolder(e["first_name"] + " " + e["last_name"]);
    createSubFolders(studentFolder,e,respondentEmail);
  }
  
  function createSubFolders(folder,e, respondentEmail){
    let fullName = e["first_name"]+" "+ e["last_name"];
  
    let satSheets = SpreadsheetApp.openById("1tA2bFuuG7KQ0-3FnjPv1MR2CnQ6LaEnbMwZCbTpfsrM").getSheets();
    let acts = DriveApp.getFileById("18DXYrair2pXqy8ZDuDkkTAJ07ynNrpqSmSqnqJkCecc");
    let blank_student_sheet = DriveApp.getFileById("1iTqTwlCbDo6guktxja2q7DgEKeQRwsmpgypfprz8K0U");
    let tpFolder = folder.createFolder("Test Prep");
    let srFolder = tpFolder.createFolder("Session Reports");
    
    /* 
    *  rather than making an empty spreadsheet and adding each sheet 1 by 1 from
    *  the sats and acts files, I just copy the ACT sheet and then add the SATs 1 by 1
    */
    let studentFile = blank_student_sheet.makeCopy("Answer Sheet - " + e["last_name"],tpFolder);
  
  
    let studentSpreadsheet = SpreadsheetApp.openById(studentFile.getId());
    enterStudentData(studentSpreadsheet,e);
    let classID = createStudentClassroom(e, respondentEmail);
  
    // put classroom ID in hidden data sheet
    Logger.log(classID);
    studentSpreadsheet.getSheetByName("data").appendRow([classID]);
    studentSpreadsheet.getSheetByName("data").appendRow([fullName]);
    studentSpreadsheet.getSheetByName("data").appendRow([folder.getId()]);
    // this is where Nathaniel will add the student ID for Firestore
  
    folder.createFolder("Other");
  }
  
  function createStudentClassroom(e, respondentEmail){
    let fname = e["first_name"];
    let lname = e["last_name"];
    let email = e["student_email"];
  
    let course = {
    "name": lname,
    "section": "Test Prep",
    "ownerId": respondentEmail,
    "guardiansEnabled": false,
  }
  
    let response = Classroom.Courses.create(course);
    Logger.log(response);
    let classID = response.id;
  
    // invite Eli and admin and Ti
    if(respondentEmail!="eli@leetutoring.com"){
      try{
          let eli = {
          "courseId": classID,
          "userId": "eli@leetutoring.com",
        }
        response = Classroom.Courses.Teachers.create(eli,classID);
      }
      catch(error){
        Logger.log("Failed to add teacher " + eli.userID + ". Error: " + error);
      };
    }
    if(respondentEmail!="edward@leetutoring.com"){
      try{
          let edward = {
          "courseId": classID,
          "userId": "edward@leetutoring.com",
        }
        response = Classroom.Courses.Teachers.create(edward,classID);
      }
      catch(error){
        Logger.log("Failed to add teacher " + edward.userID + ". Error: " + error);
      };
    }
  
    if(respondentEmail!="admin@leetutoring.com"){
      try{
        let admin = {
          "courseId": classID,
          "userId": "admin@leetutoring.com",
        }
        response = Classroom.Courses.Teachers.create(admin,classID);
      }
      catch(error){
        Logger.log("Failed to add teacher " + admin.userID + ". Error: " + error);
      };
    }
    if(respondentEmail!="ti@leetutoring.com"){
      try{
        let ti = {
          "courseId": classID,
          "userId": "ti@leetutoring.com",
        }
        response = Classroom.Courses.Teachers.create(ti,classID);
      }
      catch(error){
        Logger.log("Failed to add teacher " + ti.userID + ". Error: " + error);
      };
    }
  
    if(respondentEmail!="ben@leetutoring.com"){
      try{
        let ben = {
        "courseId": classID,
        "userId": "ben@leetutoring.com",
        }
        response = Classroom.Courses.Teachers.create(ben,classID);
      }
      catch(error){
        Logger.log("Failed to add teacher " + ben.userID + ". Error: " + error);
      };
    }
  
  
    // this doesn't work and creates bugs later so avoiding it
    /*
    let invite = {
      "userId": email,
      "courseId": classID,
      "role": "STUDENT",
    }
    response = Classroom.Invitations.create(invite);
    */
    
    return classID;
  
  }
  
  function enterStudentData(spreadsheet,e){
    // get first sheet ("Profile" sheet)
    let sheet = spreadsheet.getSheets()[0];
    sheet.getRange("B1").setValue(new Date());
    sheet.getRange("B2").setValue(e["student_email"]);
    sheet.getRange("B3").setValue(e["parent_email"]);
    sheet.getRange("B4").setValue(e["first_name"]+" "+e["last_name"]);
    sheet.getRange("B5").setValue(e["school"]);
    sheet.getRange("B6").setValue(e["grade"]);
    sheet.getRange("B7").setValue(e["test_focus"]);
    sheet.getRange("B8").setValue(e["accommodations"]);
    if(e["registered_tests"]!=null){
      sheet.getRange("B9").setValue(e["registered_tests"]);
    }
    sheet.getRange("B10").setValue(e["scheduler"]);
    sheet.getRange("B11").setValue(e["interests"]);
    sheet.getRange("B12").setValue(e["availability"]);
  }
  
  
  // SEND INFO TO BACKEND FOR DB CREATION
  function sendStudentDataToBackend(student_data) {
    let url = "https://agora-backend-1057197198698.us-east1.run.app/cmd/firestoreupdater/initializeNewStudent";
  
    // Prepare the payload
    let payload = {
      "name": student_data["first_name"] + " " + student_data["last_name"],
      "student_email": student_data["student_email"],
      "student_number": student_data["student_number"],
      "parent_email": student_data["parent_email"],
      "parent_number": student_data["parent_number"],
      "school": student_data["school"],
      "grade": student_data["grade"],
      "scheduler": student_data["scheduler"],
      "test_focus": student_data["test_focus"],
      "accommodations": student_data["accommodations"],
      "interests": student_data["interests"],
      "availability": student_data["availability"],
      "registered_for_test": student_data["registered_tests"] ? true : false,
      "test_date": student_data["registered_tests"] || null
    };
  
    // Set the options for the POST request
    let options = {
      'method': 'post',
      'contentType': 'application/json',
      'payload': JSON.stringify(payload)
    };
  
    try {
      // Send the POST request
      var response = UrlFetchApp.fetch(url, options);
      Logger.log("Data sent to backend. Response: " + response.getContentText());
    } catch (error) {
      Logger.log("Error sending data to backend: " + error);
    }
  }
  
  
  
  