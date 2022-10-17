import { useState, useEffect, useRef } from "react";
import ReactDOM from "react-dom";
import { BrowserRouter, Routes, Route, Link } from "react-router-dom";
import { useForm } from "react-hook-form";
import "./App.css";

// Components
import Registration from "./components/Registration";
import Welcome from "./components/Welcome";
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
const teams = [
  {
    name: "Qatar",
    availible: true,
  },
  {
    name: "Ecuador",
    availible: true,
  },
  {
    name: "Senegal",
    availible: false,
  },
  {
    name: "Netherlands",
    availible: true,
  },
];

console.log(teamPicker(teams));

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Welcome />} />
        <Route path="/registration" element={<Registration />} />
      </Routes>
    </BrowserRouter>
  );
}

export default App;
