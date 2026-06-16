import 'package:flutter/material.dart';
import 'services/chat_service.dart';
import 'screens/chat_screen.dart';

// Para desarrollo local con emulador Android usa http://10.0.2.2:8080
// En producción, reemplaza por la URL pública de Render (Tarea 13).
const backendUrl = String.fromEnvironment(
  'BACKEND_URL',
  defaultValue: 'http://10.0.2.2:8080',
);

void main() {
  runApp(const AsistenteApp());
}

class AsistenteApp extends StatelessWidget {
  const AsistenteApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Asistente IA',
      theme: ThemeData(colorSchemeSeed: Colors.blue, useMaterial3: true),
      home: ChatScreen(service: ChatService(baseUrl: backendUrl)),
    );
  }
}
