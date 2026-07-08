import { NestFactory } from '@nestjs/core';
import { AppModule } from './app.module';
import { ValidationPipe } from '@nestjs/common';
import helmet from 'helmet';
import { SwaggerModule, DocumentBuilder } from '@nestjs/swagger';
import { Logger } from 'nestjs-pino';
import { GlobalExceptionFilter } from './common/filters/global-exception.filter';
import { ConfigService } from '@nestjs/config';

async function bootstrap() {
  const app = await NestFactory.create(AppModule, { bufferLogs: true });
  
  // Observability & Logging
  app.useLogger(app.get(Logger));

  // Security
  app.use(helmet());
  app.enableCors();

  // Globals
  app.useGlobalPipes(new ValidationPipe({ 
    whitelist: true, 
    forbidNonWhitelisted: true,
    transform: true 
  }));
  app.useGlobalFilters(new GlobalExceptionFilter(app.get(Logger)));

  // Swagger setup
  const config = new DocumentBuilder()
    .setTitle('Aegis Auth Service')
    .setDescription('Authentication and Authorization APIs')
    .setVersion('1.0')
    .addBearerAuth()
    .build();
  const document = SwaggerModule.createDocument(app, config);
  SwaggerModule.setup('api/docs', app, document);

  // Enable graceful shutdown
  app.enableShutdownHooks();

  const configService = app.get(ConfigService);
  const port = configService.get<number>('PORT') || 4001;
  
  await app.listen(port);
  console.log(`Auth Service is running on: ${await app.getUrl()}`);
}
bootstrap();
