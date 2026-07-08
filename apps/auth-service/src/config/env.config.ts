import { z } from 'zod';

export const envValidationSchema = z.object({
  NODE_ENV: z.enum(['development', 'production', 'test']).default('development'),
  PORT: z.coerce.number().default(4001),
  DATABASE_URL: z.string().url().optional(), // Made optional for now so service can start without DB
  REDIS_URL: z.string().url().optional(),
});

export type EnvConfig = z.infer<typeof envValidationSchema>;
