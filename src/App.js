import { useState, useEffect, useRef } from "react";
import ReactDOM from "react-dom";
import { BrowserRouter, Routes, Route, Link } from "react-router-dom";
import { useForm } from "react-hook-form";
import "./App.css";

// Page Components
import Registration from "./pages/Registration";
import RegistrationSuccessful from "./pages/Registersuccessful";
import RegistrationFail from "./pages/RegisterFail";
import Login from "./pages/Login";
import Welcome from "./pages/Welcome";
import Matches from "./pages/Matches";
import Leaderboard from "./pages/Leaderboard";
import HomePageLoggedOn from "./pages/HomepageLoggedOn";
// import { db } from "./firebase-config";
// import {
//   collection,
//   getDocs,
//   addDoc,
//   updateDoc,
//   deleteDoc,
//   doc,
// } from "firebase/firestore";

//functions
const teamPicker = require("./functions/teamPicker");

//dummyData
// const teams = [
//   {
//     name: "Qatar",
//     availible: true,
//   },
//   {
//     name: "Ecuador",
//     availible: true,
//   },
//   {
//     name: "Senegal",
//     availible: false,
//   },
//   {
//     name: "Netherlands",
//     availible: true,
//   },
// ];

// console.log(teamPicker(teams));

function App() {
  return (
    <BrowserRouter>
      <Routes>
        {/* public routes */}
        <Route path="/" element={<Welcome />} />
        <Route path="/register" element={<Registration />} />
        {/* logged on routes */}
        <Route path="/success" element={<RegistrationSuccessful />} />
        <Route path="/fail" element={<RegistrationFail />} />
        <Route path="/login" element={<Login />} />
        <Route path="/home" element={<HomePageLoggedOn />} />
        <Route path="/leaderboard" element={<Leaderboard />} />
        <Route path="/matches" element={<Matches />} />
      </Routes>
    </BrowserRouter>
  );
}

export default App;
