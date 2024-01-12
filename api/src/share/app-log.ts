import { createAppLog } from "./pino-app-logger";

export const appLog = createAppLog({ level: "debug" });
