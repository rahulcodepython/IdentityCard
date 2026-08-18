# Identity Card Project Plan

This will be a identity card generator project with attendance tracker. It will help organizations to manage events, reduce the cost of printing physical identity card and track attendance. we have two types of event 

1. Stablished Event:
    A stablished event is such a event where we create an event which will last longer than one day. where users will last much longer and attend the event more than one time.

2. Flash events:
    A flash event is quick event just for a day and this event will exists only for one day and attendance will be for one day.

### Sub Events:

One event may or may not have sub events. Organizers if wants any sub-events and split peoples into multiple sub-events they can do. if one event has no sub-events then all the people of that event will be considered as whole event people. we can further assigne particular people to one or more that one sub-events selectivly. 

## How the system will work:

1. first an organization has to purchase our subscriptions.
2. then based on the plan they can create one or more than one events or sub-events.

    ### Plan:
    We will provide a very flexible plan. We have only two aspects of flash plans and stable plan. 

    1. Flash Plan: 
    Suppose I am organizing a event just for one day, I can simply purchase a plan for just one event. If I have to host another event I have to purchase another plan like a add on. The intermediate data of that flash events will stay for 30 days after the event ends.

    2. Standard Plan:
    Standard plan is for those non-flash plans people. It also has multiple aspects:

        - Very small events: For such events which may last less then 5 days.
        - Small events: For such events which may last 7 days.
        - Medium events: For such events which may last 1-2 months
        - Large events: for 6 months
        - Yearly events: last for a year
        - Recurring events: 
            - here we will check how many events they will create based on the number of events we will restrict them and charge them per year basis. we have to notify them
            - here for yearly package for unlimited evnets that will last for years. this is a custom plan we have to manually talk to us. (please build a system where we have to dynamically set the charges and gradually increases day by day as the stored data amount is increasing.)

            here we have to charge clients on yearly basis and per year the cost will raise due to storage and we can't loose this data. we have to kept them as long as the client wants.

3. now after purchasing the plan, one user admin user can create a organization and that user will be the ("Admin") of the organization.
4. admin can create events, sub-events (if needed). then add users information to an event or sub-events. the user info we will take in three forms either manually enter one by one user info by admin or upload a csv file to the server by admin or generate a public form and sends the link of that form to anyone and they can enter his details here we can have the feature of first n numbers of people (event might have the limit). the structure of the data should be (Email, mobile "unique combine", Name, "Image", "Age", "Gender"). if user selects to keep only events then we have no issues of redundant data as we just assign all the users to that event. but if admin selects the sub-event feature then they have to assign people to each events. so to do so admin might have upload the csv file in each sub-events with same user in multiple events we have to handle them as one user.  I thing this is making a little bit confusion. here per event has a separate area to collect data and insert its value. after collecting data, then comes to assign people to events and sub-events. admin can manually scroll through all the people records collected for that event and check one by one or uploads all the users info at once with csv file. admin can prior downlaod the people csv file from this platform there collects users data page. there they can select users or filter users over all the parameters and only selected once can be downloaded.

5. now for flash events admins have to set the event data start time end time, entry time exit time. based on that data the user attendence will be markded.
6. for standard events: it might be a range of dates from startin date, ending date, might vary different entry time and exit times for each day. also might have the event is happening in selective dates on a range of date. similarly suppose a event is happening from a date to a end date. now there are multiple sub-events for a selective days. 
7. now admin has to confirm the event and publish it. after publishing an identity card will be sent via email to the given email to all the people of that event. the card will be a pdf image with our company logo, event name, logo, place, date, time, to which dates and from which time to which time the identity card is permitted. also full details of that user which are being taken. and a qr code.
9. now organization admin can add other people (human) or bots in the organization whose job will be to scan and verify the qr code. there will be three roles of a organization ("Super Admin", "Admin", "Scanner") there will be only one "Super Admin", and super admin can add new people, assigne roles (one or more than one) to a user. admin user can't modify the others things just see the dashboard and handle the events. and scanner bots will be a device only (can be mobile, tab, tablet anything) that device needs to be verified first with a 6 digit otp and a encription key [like when admin has to add that bot they will simply add a bot then a otp will pop up on the screen then there will be a route on which the device has to navigate and then add the code then a encryption key will be added to the browser localstorage. with that token the device will be verified.]
10. now the event date has arrived and people are comming, any user can just hold the registered device in hand and the bot device only have one route (a scanner route. ) on that route, a button will be there calling scan in the middle of the screen if anyone clicks on it, then it will open the camera then we have to scan the qr code only of the identity cards then it will verify the id card is valid or not if valid then it will register the entry time of that user if that user is allowed to that event or sub-event. then after verification complete, it will show a simple card popup where it will show that the user is permitted to the event with person details and events details. also mention that it is his entry if this is the first scan of that day for the user. and note down that whether the user has arrived earlier for that event or late. and after event end the identity card will be scanned again to mark that he exits the event and the time whether the user has exits earlier or late.
11. after that event admins can view full analysis of the event people attented. 
    - how many people attend or absent
    - show all the people of that event or sub-event detials when arrived or not earlier or late, when exits earlier or late, 
    - also if the event or sub-event is multi date event then admin can view under a event or sub-event on which date the a user has arrival exit status with a graph of performance, and filtered by dates on which date all the users comming detials. admin can filter all the data among all the parameters and select all of few ones and download the data. also take a screenshots of graphs. we have to show the grpah per event ot sub-event among all the data.

12. for for the recurring systems like for offices, schools/universities:
    - School/university: 
        there are fixed start session or end sesion data. and also idcard validity limit
    - offices:
        evnet has no such limits also in office a person can join a event event after it started.

    we have to handle all of those edge cases

also make sure that people will also have to joined date. now for already active events, assign date will be the joining date, and for other cases admin might joined all the users but haven't published the events in that case the published date will be the joining date for all the users.

13. then we have some others page like transactions page where org super admin can only view all the transactions dowload invoice, organization settings page where admin can change the org details metadata and other config. and one dashboard with all the stats and graphs and make user aware about his plans, theri plans limitations and encourage people to purchase more. and add a page for showing all the plans. also show the landing page.



use next.js, shadcn ui, react-hook-form, zod, 
golang fiber postgress sql and redis setup. 

make sure to confirm all the coding convension how the code will be written and show the code organize, folder structure, coding happen. 
always parse form data, reqeust body data for api, resposne data after api, and other data using zod, follow strict typings.


and most important don't write redundanct code, always try to use generic code writting syle, and follow dry principle, less complex code to get better output. 

and never halusinate and started making assuptions if any doubt you have first confirm and clear the doubt from me to align my thoughts with you plan