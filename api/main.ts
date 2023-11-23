import express from "express";
import { connectToMongoDB } from "./src/utils/mongoDB";
import { startHooksListener } from "./src/services/hooksService";
import hooksRoutes from "./src/routes/hooksRoutes";
import { appConfig } from "./src/configuration";
import { appLog } from "./src/share/app-log";

const app = express();

// Connect to MongoDB
connectToMongoDB();

// Use JSON middleware
app.use(express.json());

// Use hooks routes
app.use("/sendhooks", hooksRoutes);

// Start Redis stream listener
startHooksListener();

// Start server
app.listen(appConfig.thisServer.port, () => {
  appLog.info(`Server is running on port ${appConfig.thisServer.port}`);
});
