import express from "express";
import { HookController } from "../controllers/hooksController";

const router = express.Router();

router.get("/hooks", HookController.findAll);

router.get("/hooks/:hookId", HookController.findOne);

export default router;
