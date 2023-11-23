import { nbVal, strVal } from "@paroi/data-formatters-lib";
import dotenv from "dotenv";

dotenv.config();

export const appConfig = {
  thisServer: {
    port: nbVal(process.env.PORT),
  },
  database: {
    url: strVal(process.env.MONGODB_URI),
  },
  redis: {
    host: strVal(process.env.REDIS_HOST),
    port: nbVal(process.env.REDIS_PORT),
  },
  streamName: strVal(process.env.STREAM_KEY),
};
