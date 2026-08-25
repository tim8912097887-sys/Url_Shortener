import z from "zod";

export const signupSchema = z.object({
  username: z
    .string("Username must be a string")
    .min(3, "Username must be at least 3 characters")
    .max(50, "Username must be at most 50 characters")
    .regex(/^[a-zA-Z0-9]+$/, "Username must be alphanumeric")
    .optional(),
  email: z.email().optional(),
  password: z
    .string("Password must be a string")
    .min(8, "Password must be at least 8 characters")
    .max(72, "Password must be at most 72 characters")
    .optional(),
});

export type SignupSchemaType = z.infer<typeof signupSchema>;
