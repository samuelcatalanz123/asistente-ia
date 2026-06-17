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

class AsistenteApp extends StatefulWidget {
  const AsistenteApp({super.key});

  @override
  State<AsistenteApp> createState() => _AsistenteAppState();
}

class _AsistenteAppState extends State<AsistenteApp> {
  ThemeMode _modo = ThemeMode.light;

  void _alternarTema() {
    setState(() {
      _modo = _modo == ThemeMode.light ? ThemeMode.dark : ThemeMode.light;
    });
  }

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Asistente IA',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        colorSchemeSeed: const Color(0xFF1A73E8),
        useMaterial3: true,
        brightness: Brightness.light,
      ),
      darkTheme: ThemeData(
        colorSchemeSeed: const Color(0xFF1A73E8),
        useMaterial3: true,
        brightness: Brightness.dark,
      ),
      themeMode: _modo,
      home: ChatScreen(
        service: ChatService(baseUrl: backendUrl),
        onToggleTheme: _alternarTema,
        isDark: _modo == ThemeMode.dark,
      ),
    );
  }
}
