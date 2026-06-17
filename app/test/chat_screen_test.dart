import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/testing.dart';
import 'package:http/http.dart' as http;
import 'package:asistente_ia/services/chat_service.dart';
import 'package:asistente_ia/screens/chat_screen.dart';

void main() {
  testWidgets('typing and sending shows the user bubble', (tester) async {
    // Respondemos en formato streaming (SSE), como hace el servidor real.
    final mockClient = MockClient((request) async {
      return http.Response(
        'data: {"t":"respuesta IA"}\n\ndata: [DONE]\n\n',
        200,
        headers: {'content-type': 'text/event-stream'},
      );
    });
    final service = ChatService(baseUrl: 'http://test.local', client: mockClient);

    await tester.pumpWidget(MaterialApp(
        home: ChatScreen(
      service: service,
      onToggleTheme: () {},
      isDark: false,
    )));

    await tester.enterText(find.byType(TextField), 'hola mundo');
    await tester.tap(find.byIcon(Icons.send));
    await tester.pump(); // user bubble appears immediately

    expect(find.text('hola mundo'), findsOneWidget);

    // Dejamos que lleguen los trozos del streaming (no usamos pumpAndSettle
    // porque el círculo de "cargando" gira sin parar).
    for (var i = 0; i < 5; i++) {
      await tester.pump(const Duration(milliseconds: 100));
    }
    expect(find.textContaining('respuesta IA'), findsOneWidget);
  });
}
