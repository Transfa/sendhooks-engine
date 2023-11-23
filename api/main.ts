import express from "express";
import { connectToMongoDB } from "./src/utils/mongoDB";
import { startHooksListener } from "./src/services/hooksService";
import hooksRoutes from "./src/routes/hooksRoutes";
import { appConfig } from "./src/configuration";
import { appLog } from "./src/share/app-log";

const app = express();

async function main() {
  await connectToMongoDB();

  app.use(express.json());

  app.use("/api/sendhooks/v1", hooksRoutes);

  startHooksListener();

  app.listen(appConfig.thisServer.port, () => {
    appLog.info(`Server is running on port ${appConfig.thisServer.port}`);
  });
}

main().catch(appLog.error);
