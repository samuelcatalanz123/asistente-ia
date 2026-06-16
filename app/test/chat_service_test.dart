import 'dart:convert';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:asistente_ia/models/message.dart';
import 'package:asistente_ia/services/chat_service.dart';

void main() {
  test('sendMessages returns the reply from the backend', () async {
    final mockClient = MockClient((request) async {
      expect(request.url.path, '/chat');
      final body = jsonDecode(request.body);
      expect(body['messages'][0]['content'], 'hola');
      return http.Response(jsonEncode({'reply': 'buenas'}), 200);
    });

    final service = ChatService(
      baseUrl: 'http://test.local',
      client: mockClient,
    );

    final reply = await service.sendMessages([
      const Message(role: 'user', content: 'hola'),
    ]);

    expect(reply, 'buenas');
  });

  test('sendMessages throws a friendly error on failure', () async {
    final mockClient = MockClient((request) async {
      return http.Response(jsonEncode({'error': 'boom'}), 502);
    });
    final service = ChatService(baseUrl: 'http://test.local', client: mockClient);

    await expectLater(
      service.sendMessages([const Message(role: 'user', content: 'x')]),
      throwsA(isA<Exception>()),
    );
  });
}
