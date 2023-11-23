import express from "express";
import { HookController } from "../controllers/hooksController";

const router = express.Router();

router.get("/hooks", HookController.findAll);

router.get("/hooks/:hook_id", HookController.findOne);

export default router;
