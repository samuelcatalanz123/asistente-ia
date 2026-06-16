import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/testing.dart';
import 'package:http/http.dart' as http;
import 'dart:convert';
import 'package:asistente_ia/services/chat_service.dart';
import 'package:asistente_ia/screens/chat_screen.dart';

void main() {
  testWidgets('typing and sending shows the user bubble', (tester) async {
    final mockClient = MockClient((request) async {
      return http.Response(jsonEncode({'reply': 'respuesta IA'}), 200);
    });
    final service = ChatService(baseUrl: 'http://test.local', client: mockClient);

    await tester.pumpWidget(MaterialApp(home: ChatScreen(service: service)));

    await tester.enterText(find.byType(TextField), 'hola mundo');
    await tester.tap(find.byIcon(Icons.send));
    await tester.pump(); // user bubble appears immediately

    expect(find.text('hola mundo'), findsOneWidget);

    await tester.pumpAndSettle(); // wait for async reply
    expect(find.text('respuesta IA'), findsOneWidget);
  });
}
