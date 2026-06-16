import 'package:flutter/material.dart';
import 'services/chat_service.dart';
import 'screens/chat_screen.dart';

// Por defecto apunta al servidor desplegado en Render (producción).
// Para desarrollo local puedes sobreescribirlo con:
//   flutter run --dart-define=BACKEND_URL=http://10.0.2.2:8090
const backendUrl = String.fromEnvironment(
  'BACKEND_URL',
  defaultValue: 'https://asistente-ia-xh5v.onrender.com',
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
