// utils/mongoDB.ts
import mongoose from "mongoose";
import { appConfig } from "../configuration";

export async function connectToMongoDB() {
  try {
    await mongoose.connect(appConfig.database.url);
    console.log("Connected to MongoDB");
  } catch (error) {
    console.error("Error connecting to MongoDB:", error);
    process.exit(1);
  }
}
