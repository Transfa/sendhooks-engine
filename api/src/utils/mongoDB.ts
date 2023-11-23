// utils/mongoDB.ts
import mongoose from "mongoose";
import { appConfig } from "../configuration";
import { appLog } from "../share/app-log";

export async function connectToMongoDB() {
  try {
    await mongoose.connect(appConfig.database.url);
    appLog.debug("Connected to MongoDB");
  } catch (error) {
    appLog.error("Error connecting to MongoDB:", error);
    process.exit(1);
  }
}
