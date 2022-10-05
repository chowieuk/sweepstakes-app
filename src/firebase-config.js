// Import the functions you need from the SDKs you need
import { initializeApp } from "firebase/app";
import { getAnalytics } from "firebase/analytics";
import { getAuth } from "firebase/auth";
import { getFirestore } from "@firebase/firestore";
// TODO: Add SDKs for Firebase products that you want to use
// https://firebase.google.com/docs/web/setup#available-libraries

// Your web app's Firebase configuration
// For Firebase JS SDK v7.20.0 and later, measurementId is optional
// CONVERT TO .ENV!!!
const firebaseConfig = {
  apiKey: "AIzaSyAib29tXwBpSG9z88KdbyNW4lN3cVT1yZo",
  authDomain: "sweepstakes-webapp.firebaseapp.com",
  projectId: "sweepstakes-webapp",
  storageBucket: "sweepstakes-webapp.appspot.com",
  messagingSenderId: "960797965336",
  appId: "1:960797965336:web:33d3613d47361aa0cb176d",
  measurementId: "G-EDR7KXKDW4",
};

// Initialize Firebase
const app = initializeApp(firebaseConfig);
const analytics = getAnalytics(app);
export const auth = getAuth(app);
export const db = getFirestore(app);
